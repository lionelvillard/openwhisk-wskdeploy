/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-openwhisk-client-go/whisk"
	"os"
)

/*
 * The whole purpose of this utility is to create a managed annotation for managed deployment.
 * Every OpenWhisk entity in the manifest file will be annotated with:
 * managed:
 * 	__OW__PROJECT__NAME: MyProject
 *	__OW__PROJECT_HASH: SHA1("OpenWhisk " + <size_of_manifest_file> + "\0" + <contents_of_manifest_file>)
 *	__OW__FILE: Absolute path of manifest file on file system
 */

const (
	MANAGED         = "managed"
	OPENWHISK       = "OpenWhisk"
	NULL            = "golang\000"
	OW_PROJECT_NAME = "__OW_PROJECT_NAME"
	OW_PROJECT_HASH = "__OW_PROJECT_HASH"
)

type ManagedAnnotation struct {
	ProjectName string `json:"__OW_PROJECT_NAME"`
	ProjectHash string `json:"__OW_PROJECT_HASH"`
	File        string `json:"__OW_FILE"`
}

// Project Hash is generated based on the following formula:
// SHA1("OpenWhisk " + <size_of_manifest_file> + "\0" + <contents_of_manifest_file>)
// Where the text OpenWhisk is a constant prefix
// \0 is also constant and is the NULL character
// The <size_of_manifest_file> and <contents_of_manifest_file> vary depending on the manifest file
func generateProjectHash(filePath string) (string, error) {
	projectHash := ""
	// open file to read manifest file and find its size
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return projectHash, err
	}

	// run stat on the manifest file to get its size
	f, err := file.Stat()
	if err != nil {
		return projectHash, err
	}
	size := f.Size()

	// read the file contents
	contents := make([]byte, size)
	_, err = file.Read(contents)
	if err != nil {
		return projectHash, err
	}

	// combine all the hash components used to generate SHA1
	hashContents := OPENWHISK + string(size) + NULL + string(contents)

	// generate a new hash.Hash computing the SHA1 checksum
	h := sha1.New()
	h.Write([]byte(hashContents))
	// Sum returns the SHA-1 checksum of the data
	hash := h.Sum(nil)
	// return SHA-1 checksum in hex format
	projectHash = fmt.Sprintf("%x", hash)

	return projectHash, nil
}

func GenerateManagedAnnotation(projectName string, filePath string) (whisk.KeyValue, error) {
	projectHash, err := generateProjectHash(filePath)
	managedAnnotation := whisk.KeyValue{}
	if err != nil {
		return managedAnnotation, err
	}
	m := ManagedAnnotation{
		ProjectName: projectName,
		ProjectHash: projectHash,
		File:        filePath,
	}
	ma, err := json.Marshal(m)
	if err != nil {
		return managedAnnotation, err
	}
	var a interface{}
	err = json.Unmarshal(ma, &a)
	managedAnnotation = whisk.KeyValue{Key: MANAGED, Value: a.(map[string]interface{})}
	return managedAnnotation, nil
}
