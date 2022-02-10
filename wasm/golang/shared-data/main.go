// Copyright 2020-2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
	//"io/ioutil"
	//"time"
	//"net/http"
)

const (
	sharedDataKey                 = "hello_world_shared_data_key"
)

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type (
	vmContext     struct{}
	pluginContext struct {
		// Embed the default plugin context here,
		// so that we don't need to reimplement all the methods.
		types.DefaultPluginContext
	}

	httpContext struct {
		// Embed the default http context here,
		// so that we don't need to reimplement all the methods.
		types.DefaultHttpContext
	}
)

// Override types.VMContext.
func (*vmContext) OnVMStart(vmConfigurationSize int) types.OnVMStartStatus {

	proxywasm.LogCritical("Inside Go OnVMStart")

	headers := [][2]string{
    	{":method", "GET"},
    	{":path", "/uuid"},
    	{":authority", "localhost"},
    	{":scheme", "http"},
    }

	if _, err := proxywasm.DispatchHttpCall("httpbin2", headers, nil, nil,
		5000, httpCallResponseCallback); err != nil {
		proxywasm.LogCriticalf("HttpBin2 Dispatch http call failed: %v", err)
	}

	/**
	 	http := http.Client{Timeout: time.Duration(10) * time.Second}
		resp, err := http.Get("http://SOME_URL:8001/echo?message=hello_world")
	    if err != nil {
	    	proxywasm.LogWarnf("Error calling hello_world/echo on OnVMStart: %v", err)
	    }

	    defer resp.Body.Close()

	    
	    body, err := ioutil.ReadAll(resp.Body)

	 	if err != nil {
	    	proxywasm.LogWarnf("Error parsing hello_world/echo response on OnVMStart: %v", err)
	    }


	    proxywasm.LogInfof("Response Body : %s", body)
    **/
    
    
	return types.OnVMStartStatusOK
}

func httpCallResponseCallback(numHeaders, bodySize, numTrailers int) {
		resp, _ := proxywasm.GetHttpCallResponseBody(0, bodySize)
        response_json := string(resp)
        initialValueBuf := []byte(response_json)
        if err := proxywasm.SetSharedData(sharedDataKey, initialValueBuf, 0); err != nil {
			proxywasm.LogWarnf("Error setting shared uuid data on OnVMStart: %v", err)
		}
        proxywasm.LogInfof("Httpbin2 RESPONSE %v", response_json)
}

// Override types.DefaultVMContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

// Override types.DefaultPluginContext.
func (*pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{}
}

// Override types.DefaultHttpContext.
func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	for {
		value, err := ctx.getSharedData()
		if err == nil {
			proxywasm.LogInfof("shared data value: %s", value)
		} else if errors.Is(err, types.ErrorStatusCasMismatch) {
			continue
		}
		break
	}
	return types.ActionContinue
}

func (ctx *httpContext) getSharedData() (string, error) {
	value, cas, err := proxywasm.GetSharedData(sharedDataKey)
	if err != nil {
		proxywasm.LogWarnf("Error getting shared data on OnHttpRequestHeaders with cas %d: %v ", cas, err)
		return "error", err
	}

	shared_value := string(value)
	
	return shared_value, err
}

