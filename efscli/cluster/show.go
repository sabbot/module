/*
 * Copyright (c) 2015-2018 Nexenta Systems, Inc.
 *
 * This file is part of EdgeFS Project
 * (see https://github.com/Nexenta/edgefs).
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package cluster

/*
#include "ccow.h"
*/
import "C"

import (
	"github.com/sabbot/module/efscli/efsutil"
	"github.com/sabbot/module/efscli/validate"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func Show(cpath string) error {
	return efsutil.PrintMD(cpath, "", "", "")
}

var (
	showCmd = &cobra.Command{
		Use:   "show <cluster>",
		Short: "show cluster",
		Long:  "show parameters of existing cluster",
		Args:  validate.Cluster,
		Run: func(cmd *cobra.Command, args []string) {
			err := Show(args[0])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	ClusterCmd.AddCommand(showCmd)
}
