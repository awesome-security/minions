//  Copyright 2018 Google LLC
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at

//        https://www.apache.org/licenses/LICENSE-2.0

//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//	limitations under the License.

package main

import (
	"flag"
	"log"
	"time"

	"github.com/google/minions/go/goblins"
	mpb "github.com/google/minions/proto/minions"
	pb "github.com/google/minions/proto/overlord"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	overlordAddr = flag.String("overlord_addr", "127.0.0.1:10000", "Overlord address in the format of host:port")
	rootPath     = flag.String("root_path", "/", "Root directory that we'll serve files from.")
)

func startScan(client pb.OverlordClient) []*mpb.Finding {
	log.Printf("Connecting to server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	response, err := client.CreateScan(ctx, &pb.CreateScanRequest{})
	if err != nil {
		log.Fatalf("%v.CreateScan(_) = _, %v", client, err)
	}
	scanID := response.GetScanId()
	log.Printf("Created scan %s", scanID)

	log.Printf("Will now send files for each interests, a bit at a time")

	results, err := goblins.SendFiles(client, scanID, response.GetInterests(), *rootPath)
	if err != nil {
		log.Fatalf("SendFiles %v", err)
	}
	cancel()
	return results
}

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*overlordAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewOverlordClient(conn)

	results := startScan(client)

	if len(results) == 0 {
		log.Println("Scan completed but got no vulnerabilities back. Good! Maybe.")
	}
	log.Println(goblins.HumanReadableDebug(results))
}
