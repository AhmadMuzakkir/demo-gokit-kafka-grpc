package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/ahmadmuzakkir/demo-gokit-kafka-grpc/cmd"
	"github.com/ahmadmuzakkir/demo-gokit-kafka-grpc/proto"
	"github.com/oklog/ulid"
	"github.com/segmentio/kafka-go"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

func main() {
	cliAction := newCliAction()

	app := &cli.App{
		Name:                 "cli",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{

				Name:  "grpc",
				Usage: "Commands to call the gRPC service",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "addr",
						Aliases: []string{"a"},
						Usage:   "The address of gRPC server. default is localhost:8001",
						Value:   "localhost:8001",
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:      "send",
						Usage:     "Send message",
						UsageText: "send [command options] [message]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Required: false,
								Usage:    "Message's ID. If not provided a random one will be generated",
							},
							&cli.StringFlag{
								Name:     "username",
								Aliases:  []string{"u"},
								Required: true,
								Usage:    "Message's username ",
							},
						},
						Action: cliAction.sendMessageGRPC,
					},
					{
						Name:  "get",
						Usage: "Get messages",
						Flags: []cli.Flag{
							&cli.UintFlag{
								Name:     "limit",
								Aliases:  []string{"l"},
								Required: false,
								Usage:    "Limit. default is 10",
							},
						},

						Action: cliAction.getMessageGRPC,
					},
				},
			}, {
				Name:    "kafka",
				Aliases: []string{"k"},
				Usage:   "Commands to call the kafka",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "addr",
						Aliases: []string{"a"},
						Usage:   "The address of kafka server. default is localhost:9092",
						Value:   "localhost:9092",
					},
					&cli.StringFlag{
						Name:    "topic",
						Aliases: []string{"t"},
						Usage:   "topic. default is 'message'",
						Value:   "message",
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:      "send",
						Usage:     "Send message",
						UsageText: "send [command options] [message]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Aliases:  []string{"i"},
								Required: false,
								Usage:    "Message's ID. If not provided a random one will be generated",
							},
							&cli.StringFlag{
								Name:     "username",
								Aliases:  []string{"u"},
								Required: true,
								Usage:    "Message's username ",
							},
						},
						Action: cliAction.sendMessageKafka,
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

type CliAction struct {
	entropy io.Reader
}

func newCliAction() *CliAction {
	t := time.Unix(1000000, 0)
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)

	return &CliAction{entropy: entropy}
}

func (ca *CliAction) sendMessageGRPC(c *cli.Context) error {
	addr := c.String("addr")
	if len(addr) == 0 {
		return fmt.Errorf("addr is empty")
	}

	username := c.String("username")
	if len(username) == 0 {
		return fmt.Errorf("username is empty")
	}

	var message string
	if args := c.Args().Slice(); len(args) == 0 {
		return fmt.Errorf("message is empty")
	} else {
		var sb strings.Builder
		l := len(args)

		for i := range args {
			sb.WriteString(args[i])
			if i < l-1 {
				sb.WriteString(" ")
			}
		}

		message = sb.String()
	}

	id := c.String("id")
	if len(id) == 0 {
		id = ulid.MustNew(ulid.Timestamp(time.Now()), ca.entropy).String()
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := proto.NewMessageClient(conn)

	_, err = client.Send(c.Context, &proto.MessageData{
		Id:       id,
		Msg:      message,
		Username: username,
	})
	if err != nil {
		return fmt.Errorf("Failed to send message: %w", err)
	}

	fmt.Println("Success")

	return nil
}

func (ca *CliAction) getMessageGRPC(c *cli.Context) error {
	addr := c.String("addr")
	if len(addr) == 0 {
		return fmt.Errorf("addr is empty")
	}

	limit := c.Uint("limit")
	if limit == 0 {
		limit = 10
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := proto.NewMessageClient(conn)
	res, err := client.Get(context.Background(), &proto.GetMessageRequest{Limit: int64(limit)})
	if err != nil {
		fmt.Printf("Failed to get message: %v", err)
		return nil
	}

	if len(res.Messages) == 0 {
		fmt.Println("There's no message")
		return nil
	}

	for i := range res.Messages {
		msg := res.Messages[i]
		fmt.Printf("ID: %s Username: %s Msg: %s\n", msg.Id, msg.Username, msg.Msg)
	}

	return nil
}

func (ca *CliAction) sendMessageKafka(c *cli.Context) error {
	addr := c.String("addr")
	if len(addr) == 0 {
		return fmt.Errorf("addr is empty")
	}

	username := c.String("username")
	if len(username) == 0 {
		return fmt.Errorf("username is empty")
	}

	topic := c.String("topic")
	if len(topic) == 0 {
		return fmt.Errorf("topic is empty")
	}

	// Generate random id if it's empty.
	id := c.String("id")
	if len(id) == 0 {
		id = ulid.MustNew(ulid.Timestamp(time.Now()), ca.entropy).String()
	}

	// The rest of the arguments is the message.
	var message string
	if args := c.Args().Slice(); len(args) == 0 {
		return fmt.Errorf("message is empty")
	} else {
		var sb strings.Builder
		l := len(args)

		for i := range args {
			sb.WriteString(args[i])
			if i < l-1 {
				sb.WriteString(" ")
			}
		}

		message = sb.String()
	}

	data := proto.MessageData{
		Id:       id,
		Msg:      message,
		Username: username,
	}

	payload, err := data.Marshal()
	if err != nil {
		fmt.Printf("Failed to marshal message: %v\n", err)
		return nil
	}

	if err := cmd.CreateTopic(addr, topic); err != nil {
		fmt.Printf("Failed create topic %s: %v\n", topic, err)
		return nil
	}

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{addr},
		Topic:    topic,
	})
	defer w.Close()
	fmt.Println("Sending the message...")

	err = w.WriteMessages(c.Context,
		kafka.Message{
			Key:   []byte(id),
			Value: payload,
		},
	)
	if err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
	}

	fmt.Println("Success")
	return nil
}
