package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
	"github.com/ahmadmuzakkir/demo-gokit-kafka-grpc/badgerdb"
	"github.com/ahmadmuzakkir/demo-gokit-kafka-grpc/cmd"
	"github.com/ahmadmuzakkir/demo-gokit-kafka-grpc/consumer"
	"github.com/ahmadmuzakkir/demo-gokit-kafka-grpc/inmem"
	"github.com/ahmadmuzakkir/demo-gokit-kafka-grpc/message"
	"github.com/ahmadmuzakkir/demo-gokit-kafka-grpc/proto"
	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/namsral/flag"
	"google.golang.org/grpc"
)

var (
	kafkaAddr string
	port      uint
	topic     string
	db        string
	logger    log.Logger
)

func init() {
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
}

func main() {
	flag.StringVar(&kafkaAddr, "kafka_addr", "localhost:9092", "Kafka address. default is localhost:9092")
	flag.UintVar(&port, "port", 8001, "Port for gRPC server. default is 8001")
	flag.StringVar(&topic, "topic", "message", "Kafka topic. default is 'message'")
	flag.StringVar(&db, "db", "", "Directory path to the database. If empty, a temporary in-memory database will be used instead.")
	flag.Parse()

	if kafkaAddr == "" {
		fmt.Println("kafka_addr is empty")
		os.Exit(1)
	}

	// Make sure the topic exists
	err := cmd.CreateTopic(kafkaAddr, topic)
	if err != nil {
		fmt.Printf("failed create topic %s: %v\n", topic, err)
		os.Exit(1)
	}

	// Create the repositories

	var messageRepo demo.MessageRepository
	var kafkaRepo demo.KafkaRepository

	// If db flag is empty, use in-memory db. Otherwise, use BadgerDB.
	if len(db) == 0 {
		messageRepo = inmem.NewMessageRepository()
		kafkaRepo = inmem.NewConsumerRepository()
	} else {
		badgerRepo, err := badgerdb.NewBadgerRepository(db)
		if err != nil {
			fmt.Printf("failed to create badgerdb %s: %v\n", db, err)
			os.Exit(1)
		}
		defer badgerRepo.Close()

		messageRepo = badgerRepo.MessageRepository()
		kafkaRepo = badgerRepo.KafkaRepository()
	}

	messageService := message.NewService(messageRepo, logger)

	// Starts kafka consumer

	msgConsumer := consumer.NewKafkaConsumer(consumer.KafkaConsumerConfig{
		Addr:      kafkaAddr,
		Topic:     topic,
		Partition: 0,
	}, messageService, kafkaRepo)

	go func() {
		fmt.Println("start consuming...")
		err := msgConsumer.Start()
		if err != nil {
			fmt.Println("consumer", err)
		}
	}()

	// Starts gRPC server

	grpcServer := message.NewGRPCServer(messageService)
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("failed to listen to grpc port %d: %v", port, err)
		os.Exit(1)
	}

	baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	proto.RegisterMessageServer(baseServer, grpcServer)
	go func() {
		err := baseServer.Serve(grpcListener)
		if err != nil && err != grpc.ErrServerStopped {
			fmt.Printf("grpc server: %+v\n", err)
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	<-shutdownSignal

	fmt.Println("shutting down")

	if err := msgConsumer.Stop(); err != nil {
		fmt.Printf("error stopping kafka consumer: %v\n", err)
	}

	baseServer.Stop()
}
