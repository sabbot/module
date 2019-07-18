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
package service

/*
#include "ccow.h"
*/
import "C"

import (
	"github.com/Nexenta/edgefs/src/efscli/efsutil"
	"github.com/Nexenta/edgefs/src/efscli/validate"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func Config(sname string, key string, value string) error {
	ret := efsutil.UpdateMD("", "svcs", sname, "", key, value)
	if ret != nil {
		return ret
	}
	return efsutil.PrintMDPat("", "svcs", sname, "", "X-")
}

var (
	configCmd = &cobra.Command{
		Use:   "config <service name> <key> <value>",
		Short: "configure service",
		Long:  "setup service parameter",
		Args:  validate.ServiceConfig,
		Run: func(cmd *cobra.Command, args []string) {
			err := Config(args[0], args[1], args[2])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	ServiceCmd.AddCommand(configCmd)
}
