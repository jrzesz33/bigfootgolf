package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"

	"bigfoot/golf/opsagent/internal/agent"
	"bigfoot/golf/opsagent/internal/config"
)

func main() {
	// Configure structured logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	lambda.Start(handleEventBridgeEvent)
}

func handleEventBridgeEvent(ctx context.Context, event events.CloudWatchEvent) error {
	logrus.WithFields(logrus.Fields{
		"source":      event.Source,
		"detail_type": event.DetailType,
		"event_id":    event.ID,
		"region":      event.Region,
		"account":     event.AccountID,
	}).Info("BOAT Agent: Received EventBridge event")

	// Parse the event detail
	var eventDetail map[string]interface{}
	if err := json.Unmarshal(event.Detail, &eventDetail); err != nil {
		logrus.WithError(err).Error("Failed to parse event detail")
		return fmt.Errorf("failed to parse event detail: %w", err)
	}

	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Error("Failed to load configuration")
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize the BOAT agent
	boatAgent, err := agent.NewBOATAgent(cfg)
	if err != nil {
		logrus.WithError(err).Error("Failed to initialize BOAT agent")
		return fmt.Errorf("failed to initialize BOAT agent: %w", err)
	}

	// Process the task request
	if err := boatAgent.ProcessTaskRequest(ctx, eventDetail); err != nil {
		logrus.WithError(err).Error("Failed to process task request")
		return fmt.Errorf("failed to process task request: %w", err)
	}

	logrus.Info("BOAT Agent: Successfully processed task request")
	return nil
}
