# Notes

## main() lifecycle

```
config.Load()
    database.Open()
        database.Migrate()
            repos/services
                    API Server
                        http.Server/ListenAndServe()
                            SIGTERM -> Shutdown()
                            database.Close()
```
Microgarden implementation: (?)