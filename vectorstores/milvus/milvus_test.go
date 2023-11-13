package milvus

import (
	"context"
	"os"
	"testing"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/embeddings/tei"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

func getEmbedder(t *testing.T) (embeddings.Embedder, error) {
	t.Helper()
	url := os.Getenv("TEI_URL")
	if url == "" {
		t.Skip("must set TEI_URL to run test")
	}
	return tei.New(
		tei.WithAPIBaseURL(url),
	)
}

func getNewStore(t *testing.T, opts ...Option) (Store, error) {
	t.Helper()
	url := os.Getenv("MILVUS_URL")
	if url == "" {
		t.Skip("must set MILVUS_URL to run test")
	}
	config := client.Config{
		Address: "http://localhost:19530",
	}
	e, err := getEmbedder(t)
	if err != nil {
		return Store{}, err
	}
	idx, err := entity.NewIndexAUTOINDEX(entity.L2)
	if err != nil {
		return Store{}, err
	}
	opts = append(
		opts,
		WithEmbedder(e),
		WithIndex(idx))
	return New(
		context.Background(),
		config,
		opts...,
	)
}

func TestMilvusConnection(t *testing.T) {
	t.Parallel()
	storer, err := getNewStore(t, WithDropOld())
	require.NoError(t, err)

	err = storer.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London "},
		{PageContent: "New York"},
	})
	require.NoError(t, err)
	// test with a score threshold of 0.8, expected 6 documents
	docs, err := storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0.3))
	require.NoError(t, err)
	require.Len(t, docs, 6)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0))
	require.NoError(t, err)
	require.Len(t, docs, 10)
}
