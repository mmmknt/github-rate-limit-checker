package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/pubsub"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type Rate struct {
	Limit     int `json:"limit,omitempty"`
	Remaining int `json:"remaining,omitempty"`
	Reset     int `json:"reset,omitempty"`
	Used      int `json:"used,omitempty"`
}

type RateLimit struct {
	Resources struct {
		Core                      Rate `json:"core,omitempty"`
		Graphql                   Rate `json:"graphql,omitempty"`
		Search                    Rate `json:"search,omitempty"`
		CodeSearch                Rate `json:"code_search,omitempty"`
		SourceImport              Rate `json:"source_import,omitempty"`
		IntegrationManifest       Rate `json:"integration_manifest,omitempty"`
		CodeScanningUpload        Rate `json:"code_scanning_upload,omitempty"`
		ActionsRunnerRegistration Rate `json:"actions_runner_registration,omitempty"`
		SCIM                      Rate `json:"scim,omitempty"`
		DependencySnapshot        Rate `json:"dependency_snapshot,omitempty"`
	} `json:"resources"`
	Rate Rate `json:"rate,omitempty"`
}

func getRateLimit(token string) (*RateLimit, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/rate_limit", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rateLimit RateLimit
	if err := json.NewDecoder(resp.Body).Decode(&rateLimit); err != nil {
		return nil, err
	}

	return &rateLimit, nil
}

func getGithubToken(ctx context.Context, secretName string) (string, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}

	// TODO retry?
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", err
	}
	return string(result.GetPayload().GetData()), nil
}

func main() {
	ctx := context.Background()
	// TODO secret managerの情報は、引数から取得する
	token, err := getGithubToken(ctx, "TODO")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("token:", token)

	rateLimit, err := getRateLimit(token)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	rateLimitBytes, err := json.Marshal(rateLimit)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Rate Limit: %+v\n", *rateLimit)

	psClient, err := pubsub.NewClient(ctx, "TODO set your project id")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	mstID, err := psClient.Topic("TODO set your topic id").Publish(ctx, &pubsub.Message{
		Data:       rateLimitBytes,
		Attributes: nil, // TODO
	}).Get(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Message ID:", mstID)
}
