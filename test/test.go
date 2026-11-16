package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	logger "github.com/ecpartan/soap-server-tr069/log"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/ecpartan/testing-microservice-soap/config"

	"github.com/ecpartan/soap-server-tr069/pkg/parsemap"
)

type RetScriptTask struct {
	Code    string         `json:"code"`
	Message map[string]any `json:"message"`
}

type test_context struct {
	map_inst map[string]string
}

var test_ctx test_context

func get_resp_tsk(i int, mp map[string]any) (map[string]any, error) {
	logger.LogDebug("get_resp_task", i, mp)
	if analize_task, ok := mp[strconv.Itoa(i)]; !ok {
		return nil, errors.New("Not found task1")
	} else {
		if task, ok := analize_task.(map[string]any); !ok {
			return nil, errors.New("Not found task2")
		} else {
			return task, nil
		}
	}
}

func check_code_task(task map[string]any, code string) error {

	tsk_code, ok := task["Code"]
	if !ok {
		return errors.New("Not found code")
	}

	if req_code, ok := tsk_code.(float64); ok {
		resp_port, err := strconv.Atoi(code)
		if err != nil {
			return err
		}
		if resp_port != int(req_code) {
			logger.LogDebug("tsk", task, resp_port, req_code)
			return errors.New("Code mismatch")
		}
	}
	logger.LogDebug("Code is matched")
	return nil
}

func substrInst(message string, start, end byte) (bool, int, int) {

	if idx := strings.IndexByte(message, start); idx >= 0 {
		if idx_end := strings.IndexByte(message[idx:], end); idx_end >= 0 {
			return true, idx, idx + idx_end
		} else {
			return true, idx, idx + (idx - len(message) + 1)
		}
	}

	return false, -1, -1
}

func SubstrByToken(str string, token byte, replace_map map[string]string) string {
	if ok, start, end := substrInst(str, token, '.'); ok {
		replacing := str[start:end]
		logger.LogDebug("idx", replacing)

		if replace_trim, ok := replace_map[replacing]; ok {
			return str[:start] + replace_trim + str[end:]
		}
	}

	return str
}

func add_instance(task map[string]any, message map[string]any) (error, bool) {
	if find_instance_key, ok := task["Instance"]; ok {
		inst := parsemap.GetXMLString(message, "InstanceNumber")
		logger.LogDebug("InstanceNumber22 ", reflect.TypeOf(inst), inst)
		logger.LogDebug("InstanceNumber22 ", reflect.TypeOf(find_instance_key), inst)
		if key, ok := find_instance_key.(string); ok {

			if inst != "" && find_instance_key != "" {
				test_ctx.map_inst[key] = inst
				return nil, true
			} else {
				return fmt.Errorf("Not found instance"), true
			}
		}
	}
	return nil, false
}

func check_instance(task map[string]any, message map[string]any) (error, bool) {

	if find_instance_path, ok := task["FindInstance"].(string); ok {
		logger.LogDebug("FindInstance22 ", test_ctx.map_inst)

		find_instance_path = SubstrByToken(find_instance_path, '#', test_ctx.map_inst)
		logger.LogDebug("Find instance", find_instance_path)

		paramlist := parsemap.GetXML(message, "ParameterList.ParameterValueStruct").([]any)
		find := false

		for _, v := range paramlist {
			name := parsemap.GetXMLString(v, "Name")
			if name != "" {
				find = strings.HasPrefix(name, find_instance_path)
				if find {
					logger.LogDebug("Finded ", name, find_instance_path)
					return nil, true
				}
			}
		}
		if !find {
			return fmt.Errorf("Not found name %s in return ParameterValueStruct", find_instance_path), true
		}
	}

	return nil, false
}

func find_instance(task map[string]any, message map[string]any) (error, bool) {
	logger.LogDebug("Enter ", task)

	if find_inst_values, ok := task["FindValue"].(map[string]any); ok {
		logger.LogDebug("Finded", find_inst_values)
		paramlist := make([]any, 0)
		if paramlist, ok = parsemap.GetXML(message, "ParameterList.ParameterValueStruct").([]any); !ok {
			if params, ok := parsemap.GetXML(message, "ParameterList.ParameterValueStruct").(map[string]any); ok {
				paramlist = append(paramlist, params)
			}
		}

		logger.LogDebug("find_instance", paramlist, find_inst_values)
		for path, val := range find_inst_values {
			value_mess := val.(string)
			path = SubstrByToken(path, '#', test_ctx.map_inst)
			find := false
			logger.LogDebug("Find value", path, value_mess)
			for _, v := range paramlist {
				name := parsemap.GetXMLString(v, "Name")
				if name == path {
					value_trans := parsemap.GetXMLString(v, "Value")
					if value_trans == value_mess {
						logger.LogDebug("Finded ", path, value_mess)
						return nil, true
					} else {
						return fmt.Errorf("Not found name %s with value %s in return ParameterValueStruct", path, value_mess), true
					}
				}
			}
			if !find {
				return fmt.Errorf("Not found name %s with value %s in return ParameterValueStruct", path, value_mess), true
			}

		}
	}
	return nil, false
}

func get_speed(task map[string]any, message map[string]any) (error, bool) {
	if find_inst_values, ok := task["GetSpeed"].(map[string]any); ok {
		logger.LogDebug("GetSpeed", find_inst_values)

		mode := find_inst_values["Mode"]
		logger.LogDebug("GetSpeed", mode)
		var eom, rom string
		var total int

		switch mode {
		case "Download":
			paramlist := parsemap.GetXML(message, "ParameterList.ParameterValueStruct").([]any)

			for _, v := range paramlist {
				name := parsemap.GetXMLString(v, "Name")
				if name != "" && name == "InternetGatewayDevice.DownloadDiagnostics.ROMTime" {
					rom = parsemap.GetXMLString(v, "Value")
				}

				if name != "" && name == "InternetGatewayDevice.DownloadDiagnostics.EOMTime" {
					eom = parsemap.GetXMLString(v, "Value")
				}

				if name != "" && name == "InternetGatewayDevice.DownloadDiagnostics.TotalBytesReceived" {
					totalstr := parsemap.GetXMLString(v, "Value")
					total, _ = strconv.Atoi(totalstr)
				}
			}
		case "Upload":
			paramlist := parsemap.GetXML(message, "ParameterList.ParameterValueStruct").([]any)

			for _, v := range paramlist {
				name := parsemap.GetXMLString(v, "Name")
				if name != "" && name == "InternetGatewayDevice.UploadDiagnostics.ROMTime" {
					rom = parsemap.GetXMLString(v, "Value")
				}

				if name != "" && name == "InternetGatewayDevice.UploadDiagnostics.EOMTime" {
					eom = parsemap.GetXMLString(v, "Value")
				}
				if name != "" && name == "InternetGatewayDevice.UploadDiagnostics.TotalBytesSent" {
					totalstr := parsemap.GetXMLString(v, "Value")
					total, _ = strconv.Atoi(totalstr)
				}
			}
		}
		if rom == "" || eom == "" || total == 0 {
			return fmt.Errorf("Not found rom or eom or total"), true
		}
		layout := "2006-01-02T15:04:05.999999"

		t1, _ := time.Parse(layout, rom)
		t2, _ := time.Parse(layout, eom)

		diffSeconds := t2.Sub(t1).Seconds()

		speed := (float64(total) * 8) / (1024 * 1024 * diffSeconds)

		logger.LogDebug("InstanceNumber23 ", total, rom, eom, diffSeconds, speed)

		if speed < 10 {
			return fmt.Errorf("speed < 10 Mb/s, speed = %f", speed), true
		}

		return nil, true
	}

	return nil, false
}

func updateDeviceValues(paramlist []any, mp map[string]any) map[string]any {

	for _, v := range paramlist {
		name := parsemap.GetXML(v, "Name.#text")
		value := parsemap.GetXML(v, "Value.#text")
		xtype := parsemap.GetXML(v, "Value.#attr.xsi:type.#text")
		if n, ok := name.(string); ok {
			if val, ok := value.(string); ok {
				if t, ok := xtype.(string); ok {
					repository.SetValueInJSON(mp, n, "Value", val)
					repository.SetValueInJSON(mp, n, "Type", t)
				}
			}
		}
	}

	return mp
}

func find_ipping_host(task map[string]any, message map[string]any) (error, bool) {
	logger.LogDebug("Enter ", task)

	if find_inst_values, ok := task["FindHost"].(map[string]any); ok {
		for _, val := range find_inst_values {
			if address, ok := val.(string); ok {
				logger.LogDebug("Address ", address)

				list := parsemap.GetXML(message, "ParameterList.ParameterValueStruct").([]any)
				device_cache := make(map[string]any)

				new_device_cache := updateDeviceValues(list, device_cache)
				logger.LogDebug("updateDeviceValues", new_device_cache)
				hops := parsemap.GetXML(new_device_cache, "InternetGatewayDevice.TraceRouteDiagnostics.RouteHops")
				if hops_arr := hops.(map[string]any); hops_arr != nil {
					keys := make([]int, 0, len(hops_arr))
					for k := range hops_arr {
						if kk, err := strconv.Atoi(k); err != nil {
							keys = append(keys, kk)
						} else {
							return err, true
						}
					}

					sort.Ints(keys)
					key := fmt.Sprintf("%d", keys[len(keys)-1])
					inst := parsemap.GetXMLMap(hops_arr, key)
					if inst != nil {
						logger.LogDebug("updateDeviceValues", address, parsemap.GetXMLValue(inst, "HopHostAddress"))

						if address == parsemap.GetXMLValue(inst, "HopHostAddress") {
							return nil, true
						}
						return fmt.Errorf("Addres is dissmatched"), true
					}
					return fmt.Errorf("Not found hop host"), true
				}
				return fmt.Errorf("Not found hops"), true

			}
			return fmt.Errorf("Not found address"), true
		}
		return fmt.Errorf("Not found address"), true
	}
	return nil, false
}

func config_operations(task map[string]any, message map[string]any) (error, bool) {
	logger.LogDebug("upload_conf", message)

	if _, ok := task["CheckCode"].(map[string]any); ok {
		logger.LogDebug("upload_conf", message)
		paramlist := parsemap.GetXML(message, "FaultStruct")
		str := parsemap.GetXMLString(paramlist, "FaultString")

		if str != "" {
			return fmt.Errorf("Error in upload config: %s", str), true
		}
	}
	return nil, false
}
func check_message(task map[string]any, message map[string]any) error {
	logger.LogDebug("Enter ")

	if err, ok := add_instance(task, message); ok {
		return err
	} else if err, ok := check_instance(task, message); ok {
		return err
	} else if err, ok := find_instance(task, message); ok {
		return err
	} else if err, ok := get_speed(task, message); ok {
		return err
	} else if err, ok := config_operations(task, message); ok {
		return err
	} else if err, ok := find_ipping_host(task, message); ok {
		return err
	}

	return nil
}

func AnalizeResponse(arr []RetScriptTask, path string, num string) error {
	logger.LogDebug("AnalizeResponse", "path: ", path, "num: ", num)

	cfgfw := config.GetFirmwareConfig()
	filePath := filepath.Join(cfgfw.AnalizeReportsDir, path)

	content, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer content.Close()
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	mp := make(map[string]any)

	if json.Unmarshal(data, &mp) != nil {
		return err
	}
	logger.LogDebug("Tasks: ", arr)

	script := mp[num]
	logger.LogDebug("Tasks1: ", script)

	if script == nil {
		return errors.New("script not found")
	}

	for i, rettask := range arr {

		if script_mp, ok := script.(map[string]any); ok {
			logger.LogDebug("Tasks1: ", script_mp)

			task, err := get_resp_tsk(i+1, script_mp)
			if err != nil {
				logger.LogDebug("Error: ", err)

				return err
			}
			logger.LogDebug("Task: ", task)

			if err := check_code_task(task, rettask.Code); err != nil {
				logger.LogDebug("Error: ", err)
				return err
			}

			if err := check_message(task, rettask.Message); err != nil {
				return err
			}
		}
	}

	return nil
}

func trans_script(url string, script map[string]any) (error, []RetScriptTask) {

	logger.LogDebug("script", script)
	jsonData, err := json.Marshal(script)
	if err != nil {
		return err, nil
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err, nil
	}

	client := &http.Client{
		Timeout: 300 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}

	resp.Body.Close()

	var response []RetScriptTask
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err, nil
	}

	return nil, response
}

func Execute_one_script(cfgfw *config.Firmware, filename, url string) error {

	filePath := filepath.Join(cfgfw.ScriptsDir, filename)

	content, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Error open file %s: %v\n", filePath, err)

	}
	defer content.Close()

	data, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("Error read file %s: %v\n", filePath, err)
	}

	mp := make(map[string]any)
	if json.Unmarshal(data, &mp) != nil {
		return fmt.Errorf("JSON err: %v", err)
	}

	keys := make([]string, 0, len(mp))
	for k := range mp {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {

		if script, ok := mp[k]; ok {
			if script_mp, ok := script.(map[string]any); ok {
				if curr, ok := script_mp["Script"].(map[string]any); ok {
					curr["Serial"] = cfgfw.Serial
				}
				err, response := trans_script(url, script_mp)
				if err != nil {
					return fmt.Errorf("Error trans script %s: %v\n", filePath, err)
				}
				if err = AnalizeResponse(response, filename, k); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("Not found key %s in file %s\n", k, filePath)
			}
		} else {
			return fmt.Errorf("Not found keys in file %s\n", filePath)
		}
	}

	return nil

}

func Execute_test(cfg *config.ServerInfo, cfgfw *config.Firmware, url string) map[string]string {

	files, err := os.ReadDir(cfgfw.ScriptsDir)

	test_ctx = test_context{map_inst: make(map[string]string)}
	report := make(map[string]string)

	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		err := Execute_one_script(cfgfw, file.Name(), url)

		file_key := strings.Split(file.Name(), ".")[0]
		if err != nil {
			report[file_key] = "FAILED!!! " + err.Error()
		} else {
			report[file_key] = "SUCCESS"
		}
	}

	return report
}
