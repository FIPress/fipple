# fipple

[fipple](https://fipress.org/project/fipple) is a RESTful web service framework in Go.

**Usage**

To set up a RESTful web service with `fipple`, you just need following 2 steps.
1. Register services
```
fipple.Get("/:id",function(ctx fipple.Context) {
    //a GET service
    id := ctx.GetStringParam("id")
})

fipple.Post("/create",function(ctx fipple.Context) {
    //a POST service
})
```

2. Start
```
fipple.Start(port)
```

**Context**
Context is your friend. It provides approach to both `Request` and `Response`.

For example, to get parameters:
```
ctx.GetStringParam("id")
``` 

To get posted data:
```
user := new(User)
err := ctx.GetEntity(user)
```

Or, to send response:
- Status
```
ctx.ServeStatus( http.StatusOK)
```

- JSON data
```
ctx.ServeJson(user)
```

- HTML

- XML

- Plain text

- HTML file

- Static file

- By template

For detailed usage, please refer to the [project page](https://fipress.org/project/fipple)