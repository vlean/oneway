export default {
  'GET /api/forward/list': (req: any, res: any) => {
    res.json({
      code: 0,
      data: [
        {
            "id": 1,
            "created_at": "0001-01-01T00:00:00Z",
            "updated_at": "0001-01-01T00:00:00Z",
            "deleted_at": null,
            "from": "max.demo.mock",
            "to": "github.com",
            "schema": "https",
            "client": "default",
            "rewrite":{
                "host_rewrite": true, 
                "rule":[
                    {
                        "path": "all", 
                        "request": [
                            {"type": "header", "from": "", "to": ""},
                            {"type": "header", "action":"remove", "from": "Cookie", "to": ""}
                        ],
                        "response": [
                            {"type": "body", "from": "", "to": ""}
                        ]
                    }
                ]
            },
            "status": 1
        },
        {
            "id": 2,
            "created_at": "0001-01-01T00:00:00Z",
            "updated_at": "0001-01-01T00:00:00Z",
            "deleted_at": null,
            "from": "baidu.demo.mock",
            "to": "www.baidu.com",
            "schema": "https",
            "client": "default",
            "status": 1
        },
        {
            "id": 3,
            "created_at": "0001-01-01T00:00:00Z",
            "updated_at": "2023-11-03T13:49:00.700271337+08:00",
            "deleted_at": null,
            "from": "tank.demo.mock",
            "to": "www.google.com",
            "schema": "http",
            "client": "default",
            "status": 1
        },
        {
            "id": 4,
            "created_at": "0001-01-01T00:00:00Z",
            "updated_at": "2023-11-02T15:09:57.064047489+08:00",
            "deleted_at": null,
            "from": "dash.demo.mock",
            "to": "t.cn:4000",
            "schema": "http",
            "client": "default",
            "status": 1
        }
    ],
      msg: "succ"
    });
  },
  'GET /api/stat/info': {
    "code": 0,
    "data": {
        "client": [
            {
                "name": "codespaces-219220",
                "size": 1,
                "use": 1,
                "start_at": "2023-11-03 14:00:42",
                "run_time": 800
            },
            {
                "name": "default",
                "size": 0,
                "use": 0
            }
        ],
        "http": {
            "request": 52,
            "body_size": 0,
            "auth_fail": 18
        },
        "run_time": 856.623148233,
        "start_at": "2023-11-03 14:00:42"
    },
    "msg": ""
  },
  'POST /api/forward/save': {
    code: 0,
    msg: "succ"
  },
  'POST /api/system/config/update': {
    code: 0,
    msg: 'succ'
  },
  'GET /api/system/config': {
    "code": 0,
    "data": {
        "System": {
              "Host": "0.0.0.0",
              "Port": 443,
              "Domain": "xxxx",
              "Mode": "strict"
          },
          "Client": {
              "Remote": "xxxx",
              "Name": "xxxx"
          },
          "Cloudflare": {
              "Email": "xxx@qq.com",
              "ApiKey": "xxxx",
              "DnsApiToken": "xxx",
              "ZoneApiToken": "xxx"
          },
          "Auth": {
              "Mode": "gitee",
              "Expire": 24,
              "Email": [
                  "xxx@qq.com", "demo@qq.com"
              ],
              "Token": "xxx",
              "ClientId": "xxx"
          },
          "Server": {
              "Domain": "",
              "ForceHttps": true
          }
      },
      "msg": ""
  },
};
