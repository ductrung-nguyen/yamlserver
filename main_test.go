package main

import (
    "net/http"
    "net/url"
    "reflect"
    "testing"
)

func Test_checkQuery(t *testing.T) {
    type args struct {
        query     url.Values
        condQuery map[string]interface{}
    }
    tests := []struct {
        name string
        args args
        want bool
    }{
        {
            name: "test 1: matching ok",
            args: args{
                query: url.Values{
                    "time": []string{"10"},
                    "age":  []string{"30"},
                },
                condQuery: map[string]interface{}{
                    "time": 10,
                    "age":  30,
                },
            },
            want: true,
        },
        {
            name: "test 2: doesn't match",
            args: args{
                query: url.Values{
                    "time": []string{"10"},
                    "age":  []string{"30"},
                },
                condQuery: map[string]interface{}{
                    "time": 10,
                    "age":  40,
                },
            },
            want: false,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := checkQuery(tt.args.query, tt.args.condQuery); got != tt.want {
                t.Errorf("checkQuery() = %v, want %v", got, tt.want)
            }
        })
    }
}

func Test_checkPayload(t *testing.T) {
    type args struct {
        condPayload map[string]interface{}
        realPayload map[string]interface{}
    }
    tests := []struct {
        name string
        args args
        want bool
    }{
        {
            name: "test 1: simple matching ok",
            args: args{
                realPayload: map[string]interface{}{
                    "username": "nguyend",
                    "password": "test",
                    "data":     10,
                },
                condPayload: map[string]interface{}{
                    "data":     10,
                    "username": "nguyend",
                },
            },
            want: true,
        },
        {
            name: "test 2: nested structure matching ok",
            args: args{
                realPayload: map[string]interface{}{
                    "username": "nguyend",
                    "password": "test",
                    "data": map[string]interface{}{
                        "age":     10,
                        "score":   1.2,
                        "address": "France",
                    },
                },
                condPayload: map[string]interface{}{
                    "username": "nguyend",
                    "data": map[string]interface{}{
                        "age":     10,
                        "address": "France",
                    },
                },
            },
            want: true,
        },
        {
            name: "test 3: nested structure matching failed",
            args: args{
                realPayload: map[string]interface{}{
                    "username": "nguyend",
                    "password": "test",
                    "data": map[string]interface{}{
                        "age":     10,
                        "score":   1.2,
                        "address": "France",
                    },
                },
                condPayload: map[string]interface{}{
                    "username": "nguyend",
                    "password": "test",
                    "data": map[string]interface{}{
                        "age":     10,
                        "score":   1.2,
                        "address": "Italy",
                    },
                },
            },
            want: false,
        },
        {
            name: "test 4: nested structure matching ok",
            args: args{
                realPayload: map[string]interface{}{
                    "username": "nguyend",
                    "password": "test",
                    "data": map[string]interface{}{
                        "age":   10,
                        "score": 1.2,
                        "address": map[string]interface{}{
                            "primary":   "France",
                            "secondary": "Italy",
                        },
                    },
                },
                condPayload: map[string]interface{}{
                    "username": "nguyend",
                    "password": "test",
                    "data": map[string]interface{}{
                        "age":   10,
                        "score": 1.2,
                        "address": map[string]interface{}{
                            "primary":   "France",
                            "secondary": "Italy",
                        },
                    },
                },
            },
            want: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := checkPayload(tt.args.realPayload, tt.args.condPayload); got != tt.want {
                t.Errorf("checkPayload() = %v, want %v", got, tt.want)
            }
        })
    }
}

func Test_checkConditionAndReturn(t *testing.T) {
    type args struct {
        query      url.Values
        payload    map[string]interface{}
        header     http.Header
        willReturn []Result
    }
    tests := []struct {
        name string
        args args
        want *Response
    }{
        {
            name: "test 1: simple case with query only",
            args: args{
                query: url.Values{
                    "time": []string{"10"},
                    "age":  []string{"30"},
                },
                payload: nil,
                header:  nil,
                willReturn: []Result{
                    {
                        When: &Condition{
                            Query: map[string]interface{}{
                                "time": 10,
                                "age":  30,
                            },
                        },
                        Response: Response{
                            ReturnCode:   200,
                            ReturnObject: []string{"1", "2", "3"},
                        },
                    },
                    {
                        When: &Condition{
                            Query: map[string]interface{}{
                                "time": 10,
                                "age":  40,
                            },
                        },
                        Response: Response{
                            ReturnCode:   200,
                            ReturnObject: []string{"4", "5", "6"},
                        },
                    },
                },
            },
            want: &Response{
                ReturnCode:   200,
                ReturnObject: []string{"1", "2", "3"},
            },
        },
        {
            name: "test 2: simple case with query only - doesn't match",
            args: args{
                query: url.Values{
                    "time": []string{"10"},
                    "age":  []string{"30"},
                },
                payload: nil,
                header:  nil,
                willReturn: []Result{
                    {
                        When: &Condition{
                            Query: map[string]interface{}{
                                "time": 10,
                                "age":  35,
                            },
                        },
                        Response: Response{
                            ReturnCode:   200,
                            ReturnObject: []string{"1", "2", "3"},
                        },
                    },
                    {
                        When: &Condition{
                            Query: map[string]interface{}{
                                "time": 10,
                                "age":  40,
                            },
                        },
                        Response: Response{
                            ReturnCode:   200,
                            ReturnObject: []string{"4", "5", "6"},
                        },
                    },
                },
            },
            want: nil,
        },
        {
            name: "test 3: simple case with payload only - ok",
            args: args{
                query: nil,
                payload: map[string]interface{}{
                    "time": 10,
                    "age":  30,
                },
                header: nil,
                willReturn: []Result{
                    {
                        When: &Condition{
                            Payload: map[string]interface{}{
                                "time": 10,
                                "age":  30,
                            },
                        },
                        Response: Response{
                            ReturnCode:   200,
                            ReturnObject: []string{"1", "2", "3"},
                        },
                    },
                    {
                        When: &Condition{
                            Payload: map[string]interface{}{
                                "time": 10,
                                "age":  40,
                            },
                        },
                        Response: Response{
                            ReturnCode:   200,
                            ReturnObject: []string{"4", "5", "6"},
                        },
                    },
                },
            },
            want: &Response{
                ReturnCode:   200,
                ReturnObject: []string{"1", "2", "3"},
            },
        },
        {
            name: "test 4: simple case with payload only - doesn't match",
            args: args{
                query: nil,
                payload: map[string]interface{}{
                    "time": 10,
                    "age":  30,
                },
                header: nil,
                willReturn: []Result{
                    {
                        When: &Condition{
                            Payload: map[string]interface{}{
                                "time": 10,
                                "age":  35,
                            },
                        },
                        Response: Response{
                            ReturnCode:   200,
                            ReturnObject: []string{"1", "2", "3"},
                        },
                    },
                    {
                        When: &Condition{
                            Payload: map[string]interface{}{
                                "time": 10,
                                "age":  40,
                            },
                        },
                        Response: Response{
                            ReturnCode:   200,
                            ReturnObject: []string{"4", "5", "6"},
                        },
                    },
                },
            },
            want: nil,
        },
        {
            name: "test 5: nested payload - ok",
            args: args{
                query: nil,
                payload: map[string]interface{}{
                    "username": "nguyend",
                    "password": "test",
                    "data": map[string]interface{}{
                        "age":   10,
                        "score": 1.2,
                        "address": map[string]interface{}{
                            "primary":   "France",
                            "secondary": "Italy",
                        },
                    },
                },
                header: nil,
                willReturn: []Result{
                    {
                        When: &Condition{
                            Payload: map[string]interface{}{
                                "time": 10,
                                "age":  30,
                            },
                        },
                        Response: Response{
                            ReturnCode:   200,
                            ReturnObject: []string{"1", "2", "3"},
                        },
                    },
                    {
                        When: &Condition{
                            Payload: map[string]interface{}{
                                "data": map[string]interface{}{
                                    "address": map[string]interface{}{
                                        "primary": "France",
                                    },
                                },
                            },
                        },
                        Response: Response{
                            ReturnCode:   200,
                            ReturnObject: []string{"4", "5", "6"},
                        },
                    },
                },
            },
            want: &Response{
                ReturnCode:   200,
                ReturnObject: []string{"4", "5", "6"},
            },
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := checkConditionAndReturn(tt.args.query, tt.args.payload, tt.args.header, tt.args.willReturn); !reflect.DeepEqual(got, tt.want) {
                t.Errorf("checkConditionAndReturn() = %v, want %v", got, tt.want)
            }
        })
    }
}
