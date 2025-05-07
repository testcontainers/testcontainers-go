package qdrant_test

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/qdrant"
)

func ExampleRun() {
	// runQdrantContainer {
	ctx := context.Background()

	qdrantContainer, err := qdrant.Run(ctx, "qdrant/qdrant:v1.7.4")
	defer func() {
		if err := testcontainers.TerminateContainer(qdrantContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := qdrantContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_createPoints() {
	// fullExample {
	qdrantContainer, err := qdrant.Run(context.Background(), "qdrant/qdrant:v1.7.4")
	defer func() {
		if err := testcontainers.TerminateContainer(qdrantContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	grpcEndpoint, err := qdrantContainer.GRPCEndpoint(context.Background())
	if err != nil {
		log.Printf("failed to get gRPC endpoint: %s", err)
		return
	}

	// Set up a connection to the server.
	conn, err := grpc.NewClient(grpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
		return
	}
	defer conn.Close()

	collectionsClient := pb.NewCollectionsClient(conn)

	const (
		collectionName        = "test_collection"
		vectorSize     uint64 = 4
		distance              = pb.Distance_Dot
	)

	// 1. create the collection
	var defaultSegmentNumber uint64 = 2
	_, err = collectionsClient.Create(context.Background(), &pb.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: &pb.VectorsConfig{Config: &pb.VectorsConfig_Params{
			Params: &pb.VectorParams{
				Size:     vectorSize,
				Distance: distance,
			},
		}},
		OptimizersConfig: &pb.OptimizersConfigDiff{
			DefaultSegmentNumber: &defaultSegmentNumber,
		},
	})
	if err != nil {
		log.Printf("Could not create collection: %v", err)
		return
	}

	// 2. Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := collectionsClient.List(ctx, &pb.ListCollectionsRequest{})
	if err != nil {
		log.Printf("could not get collections: %v", err)
		return
	}
	fmt.Printf("List of collections: %s\n", r.GetCollections())

	// 3. Create points grpc client
	pointsClient := pb.NewPointsClient(conn)

	// 4. Create keyword field index
	fieldIndex1Type := pb.FieldType_FieldTypeKeyword
	fieldIndex1Name := "city"
	_, err = pointsClient.CreateFieldIndex(context.Background(), &pb.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      fieldIndex1Name,
		FieldType:      &fieldIndex1Type,
	})
	if err != nil {
		log.Printf("Could not create field index: %v", err)
		return
	}

	// 5. Create integer field index
	fieldIndex2Type := pb.FieldType_FieldTypeInteger
	fieldIndex2Name := "count"
	_, err = pointsClient.CreateFieldIndex(context.Background(), &pb.CreateFieldIndexCollection{
		CollectionName: collectionName,
		FieldName:      fieldIndex2Name,
		FieldType:      &fieldIndex2Type,
	})
	if err != nil {
		log.Printf("Could not create field index: %v", err)
		return
	}

	// 6. Upsert points
	waitUpsert := true
	upsertPoints := []*pb.PointStruct{
		{
			// Point Id is number or UUID
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 1},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.05, 0.61, 0.76, 0.74}}}},
			Payload: map[string]*pb.Value{
				"city": {
					Kind: &pb.Value_StringValue{StringValue: "Berlin"},
				},
				"country": {
					Kind: &pb.Value_StringValue{StringValue: "Germany"},
				},
				"count": {
					Kind: &pb.Value_IntegerValue{IntegerValue: 1000000},
				},
				"square": {
					Kind: &pb.Value_DoubleValue{DoubleValue: 12.5},
				},
			},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 2},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.19, 0.81, 0.75, 0.11}}}},
			Payload: map[string]*pb.Value{
				"city": {
					Kind: &pb.Value_ListValue{
						ListValue: &pb.ListValue{
							Values: []*pb.Value{
								{
									Kind: &pb.Value_StringValue{StringValue: "Berlin"},
								},
								{
									Kind: &pb.Value_StringValue{StringValue: "London"},
								},
							},
						},
					},
				},
			},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 3},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.36, 0.55, 0.47, 0.94}}}},
			Payload: map[string]*pb.Value{
				"city": {
					Kind: &pb.Value_ListValue{
						ListValue: &pb.ListValue{
							Values: []*pb.Value{
								{
									Kind: &pb.Value_StringValue{StringValue: "Berlin"},
								},
								{
									Kind: &pb.Value_StringValue{StringValue: "Moscow"},
								},
							},
						},
					},
				},
			},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 4},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.18, 0.01, 0.85, 0.80}}}},
			Payload: map[string]*pb.Value{
				"city": {
					Kind: &pb.Value_ListValue{
						ListValue: &pb.ListValue{
							Values: []*pb.Value{
								{
									Kind: &pb.Value_StringValue{StringValue: "London"},
								},
								{
									Kind: &pb.Value_StringValue{StringValue: "Moscow"},
								},
							},
						},
					},
				},
			},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 5},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.24, 0.18, 0.22, 0.44}}}},
			Payload: map[string]*pb.Value{
				"count": {
					Kind: &pb.Value_ListValue{
						ListValue: &pb.ListValue{
							Values: []*pb.Value{
								{
									Kind: &pb.Value_IntegerValue{IntegerValue: 0},
								},
							},
						},
					},
				},
			},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Num{Num: 6},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.35, 0.08, 0.11, 0.44}}}},
			Payload: map[string]*pb.Value{},
		},
		{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{Uuid: "58384991-3295-4e21-b711-fd3b94fa73e3"},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: []float32{0.35, 0.08, 0.11, 0.44}}}},
			Payload: map[string]*pb.Value{},
		},
	}
	_, err = pointsClient.Upsert(context.Background(), &pb.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &waitUpsert,
		Points:         upsertPoints,
	})
	if err != nil {
		log.Printf("Could not upsert points: %v", err)
		return
	}

	// 7. Retrieve points by ids
	pointsByID, err := pointsClient.Get(context.Background(), &pb.GetPoints{
		CollectionName: collectionName,
		Ids: []*pb.PointId{
			{PointIdOptions: &pb.PointId_Num{Num: 1}},
			{PointIdOptions: &pb.PointId_Num{Num: 2}},
		},
	})
	if err != nil {
		log.Printf("Could not retrieve points: %v", err)
		return
	}

	fmt.Printf("Retrieved points: %d\n", len(pointsByID.GetResult()))

	// 8. Unfiltered search
	unfilteredSearchResult, err := pointsClient.Search(context.Background(), &pb.SearchPoints{
		CollectionName: collectionName,
		Vector:         []float32{0.2, 0.1, 0.9, 0.7},
		Limit:          3,
		// Include all payload and vectors in the search result
		WithVectors: &pb.WithVectorsSelector{SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: true}},
		WithPayload: &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		log.Printf("Could not search points: %v", err)
		return
	}

	fmt.Printf("Found points: %d\n", len(unfilteredSearchResult.GetResult()))

	// 9. filtered search
	filteredSearchResult, err := pointsClient.Search(ctx, &pb.SearchPoints{
		CollectionName: collectionName,
		Vector:         []float32{0.2, 0.1, 0.9, 0.7},
		Limit:          3,
		Filter: &pb.Filter{
			Should: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "city",
							Match: &pb.Match{
								MatchValue: &pb.Match_Keyword{
									Keyword: "London",
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("Could not search points: %v", err)
		return
	}
	// }

	fmt.Printf("Found points: %d\n", len(filteredSearchResult.GetResult()))

	// Output:
	// List of collections: [name:"test_collection"]
	// Retrieved points: 2
	// Found points: 3
	// Found points: 2
}
