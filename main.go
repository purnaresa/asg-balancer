package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"

	log "github.com/sirupsen/logrus"
)

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		err := setUnhealthy(message.Body)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}

func setUnhealthy(instanceId string) (err error) {

	// AWS Configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(retry.NewStandard(), 5)
		}),
		config.WithDefaultRegion(os.Getenv("region")))
	if err != nil {
		log.Errorln(err)
		return
	}
	//

	asgClient := autoscaling.NewFromConfig(cfg)

	inputHealth := autoscaling.SetInstanceHealthInput{
		InstanceId:   aws.String(instanceId),
		HealthStatus: aws.String("Unhealthy"),
	}
	_, err = asgClient.SetInstanceHealth(context.Background(),
		&inputHealth)
	if err != nil {
		log.Errorln(err)
		return
	}
	log.WithField("instance ID", instanceId).
		Infoln("set instance health success")
	return
}
