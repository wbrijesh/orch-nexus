package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sigv4-auth-cassandra-gocql-driver-plugin/sigv4"
	"github.com/gocql/gocql"
	_ "github.com/joho/godotenv/autoload"
)

// Service defines the interface for health checks.
type Service interface {
	Health() map[string]string
	Close() error
	InitSchema() error
}

// service implements the Service interface.
type service struct {
	Session *gocql.Session
}

// Environment variables for Amazon Keyspaces connection
var (
	endpoint = os.Getenv("KEYSPACES_ENDPOINT") // The Amazon Keyspaces endpoint
	certPath = os.Getenv("CERT_PATH")          // Path to the Starfield certificate
	// AWS credentials are automatically picked up from environment variables:
	// AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_DEFAULT_REGION
)

// New initializes a new Service with an Amazon Keyspaces Session.
func New() Service {
	cluster := gocql.NewCluster(endpoint)
	cluster.Port = 9142

	// Set up SSL/TLS
	cluster.SslOpts = &gocql.SslOptions{
		CaPath:                 certPath,
		EnableHostVerification: false,
	}

	// Use SigV4 authentication
	cluster.Authenticator = sigv4.NewAwsAuthenticator()

	// Amazon Keyspaces requires at least LOCAL_QUORUM consistency
	cluster.Consistency = gocql.LocalQuorum
	cluster.DisableInitialHostLookup = false

	// Create Session
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to connect to Amazon Keyspaces: %v", err)
	}

	s := &service{Session: session}
	return s
}

// Health returns the health status and statistics of the Amazon Keyspaces connection.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats := make(map[string]string)
	startedAt := time.Now()

	// Amazon Keyspaces doesn't support the now() function, so we'll use a simpler query
	// that is supported to check connectivity
	query := "SELECT release_version FROM system.local"
	iter := s.Session.Query(query).WithContext(ctx).Iter()
	var version string
	if !iter.Scan(&version) {
		if err := iter.Close(); err != nil {
			stats["status"] = "down"
			stats["message"] = fmt.Sprintf("Failed to execute query: %v", err)
			return stats
		}
	}

	if err := iter.Close(); err != nil {
		stats["status"] = "down"
		stats["message"] = fmt.Sprintf("Error during query execution: %v", err)
		return stats
	}

	// System is up
	stats["status"] = "up"
	stats["message"] = "It's healthy"
	stats["keyspaces_version"] = version
	stats["current_time"] = time.Now().String() // Use Go's time instead of DB time

	// Get keyspace information - this query is supported by Amazon Keyspaces
	getKeyspacesQuery := "SELECT keyspace_name FROM system_schema.keyspaces"
	keyspacesIterator := s.Session.Query(getKeyspacesQuery).Iter()

	stats["keyspaces_count"] = strconv.Itoa(keyspacesIterator.NumRows())

	// Also get the keyspace names
	var keyspaceNames []string
	var keyspaceName string
	for keyspacesIterator.Scan(&keyspaceName) {
		keyspaceNames = append(keyspaceNames, keyspaceName)
	}

	if err := keyspacesIterator.Close(); err != nil {
		log.Printf("Failed to close keyspaces iterator: %v", err)
	}

	// Get region information
	stats["aws_region"] = os.Getenv("AWS_DEFAULT_REGION")
	stats["endpoint"] = endpoint

	// Calculate the time taken to perform the health check
	stats["health_check_duration"] = time.Since(startedAt).String()
	return stats
}

// Close gracefully closes the Keyspaces Session.
func (s *service) Close() error {
	s.Session.Close()
	return nil
}
