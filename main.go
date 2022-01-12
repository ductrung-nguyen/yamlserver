package main

import (
    "flag"
    "fmt"
    "net/http"
    "net/url"
    "os"
    "path"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/kardianos/osext"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Server struct {
        Host string `yaml:"host"`
        Port int    `yaml:"port"`
    } `yaml:"server"`
    Endpoints []Endpoint `yaml:"endpoints"`
}

type Condition struct {
    Query   map[string]interface{} `yaml:"query,omitempty"`
    Payload map[string]interface{} `yaml:"payload,omitempty"`
    Header  map[string]interface{} `yaml:"header,omitempty"`
}

type Response struct {
    ReturnCode   int         `yaml:"returnCode"`
    ReturnObject interface{} `yaml:"returnObject"`
}
type Result struct {
    When     *Condition `yaml:"when,omitempty"`
    Response Response   `yaml:"response"`
}
type Endpoint struct {
    Path    string   `yaml:"path"`
    Method  string   `yaml:"method"`
    Results []Result `yaml:"results"`
}

// NewConfig reads a file in a given path and return the pointer to the configuration object
func newConfig(configPath string) (*Config, error) {
    // create a defaul configuration
    config := &Config{}

    file, err := os.Open(configPath)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    fmt.Printf("Read configuration from '%s'\n", configPath)

    // Init a new YAML decoder
    d := yaml.NewDecoder(file)

    // start YAML decoding from file
    if err := d.Decode(config); err != nil {
        return nil, err
    }

    return config, nil

}

// LoadConfigFile returns the configuration by reading the config file in the current directory
func LoadConfigFile() *Config {
    curDir, _ := osext.ExecutableFolder()
    if curDir == "" {
        // if we can not get the current executable folder, use the current directory
        curDir, _ = os.Getwd()
    }

    configPath := path.Join(curDir, "config.yaml")
    return LoadConfigFileFromPath(configPath)
}

// LoadConfigFileFromPath returns the configuration by reading the config file in the current directory
func LoadConfigFileFromPath(filepath string) *Config {
    var curDir string
    if !strings.HasPrefix(filepath, "/") {
        curDir, _ = osext.ExecutableFolder()
        if curDir == "" {
            // if we can not get the current executable folder, use the current directory
            curDir, _ = os.Getwd()
        }
    }

    configPath := path.Join(curDir, filepath)
    config, err := newConfig(configPath)
    if err != nil {
        fmt.Printf("Error: Could not load the configration from '%s':\n %s\n", configPath, err)
        return nil
    }

    return config
}

type ginMethodFunction func(string, ...gin.HandlerFunc) gin.IRoutes

func getMethodFunction(r *gin.Engine, method string) ginMethodFunction {
    switch method {
    case http.MethodGet:
        return r.GET
    case http.MethodPost:
        return r.POST
    case http.MethodPut:
        return r.PUT
    case http.MethodDelete:
        return r.DELETE
    case http.MethodPatch:
        return r.PATCH
    }
    return r.Any
}

// checkQuery returns true if the conditions in condQuery match the URL query in the request
func checkQuery(query url.Values, condQuery map[string]interface{}) bool {
    // check the URL query
    for condQueryKey, condQueryValue := range condQuery {
        queryValue := query.Get(condQueryKey)
        if queryValue != fmt.Sprint(condQueryValue) {
            return false
        }
    }
    return true
}

// checkHeader returns true if the conditions in condHeader match the header in the request
func checkHeader(header http.Header, condHeader map[string]interface{}) bool {
    return checkQuery(url.Values(header), condHeader)
}

// checkPayload returns true if the conditions in condPayload match the realPayload
func checkPayload(realPayload map[string]interface{}, condPayload map[string]interface{}) bool {
    for condPayloadKey, condPayloadValue := range condPayload {
        switch v := condPayloadValue.(type) {
        case int, int32, uint, int64, uint32, uint64, float32, float64:
            if v != realPayload[condPayloadKey] {
                return false
            }
        case string:
            if fmt.Sprint(v) != fmt.Sprint(realPayload[condPayloadKey]) {
                return false
            }
        case map[string]interface{}:
            castedMap, ok := realPayload[condPayloadKey].(map[string]interface{})
            if !ok || !checkPayload(castedMap, v) {
                return false
            }
        }
    }
    return true
}

// scan through all posibility and check the condition of each of them
// if the condition matches, return the corresponding response
func checkConditionAndReturn(query url.Values, payload map[string]interface{}, header http.Header, willReturn []Result) *Response {
    for _, retWhen := range willReturn {
        if retWhen.When != nil {
            if !checkQuery(query, retWhen.When.Query) || !checkPayload(payload, retWhen.When.Payload) || !checkHeader(header, retWhen.When.Header) {
                continue
            }
        }

        return &retWhen.Response
    }
    return nil
}

// setUpRoute registers the route and the handler function with gin based on the configuration file
func setUpRoute(r *gin.Engine, route Endpoint) {
    methodFunc := getMethodFunction(r, route.Method)

    methodFunc(route.Path, func(c *gin.Context) {
        // return the response according to the route condition
        payload := map[string]interface{}{}
        c.ShouldBindJSON(&payload)
        query := c.Request.URL.Query()
        header := c.Request.Header

        matchedResponse := checkConditionAndReturn(query, payload, header, route.Results)
        if matchedResponse != nil {
            c.JSON(matchedResponse.ReturnCode, matchedResponse.ReturnObject)
            return
        } else {
            c.Status(http.StatusNotFound)
            return
        }
    })
}

func main() {
    configPath := flag.String("config", "config.yaml", "Path to the custom config file. Keep empty to use the `config.yaml` in the current folder")
    flag.Parse()

    r := gin.Default()

    config := LoadConfigFileFromPath(*configPath)
    if config == nil {
        panic("Could not load the configuration file")
    }

    for _, route := range config.Endpoints {
        setUpRoute(r, route)
    }

    r.Run(fmt.Sprintf("%s:%v", config.Server.Host, config.Server.Port))
}
