```bash
bin                 - contains built naavik binaries
build/
└── docker          - contains dockerfiles
cmd                 - contains main.go files and start up args initialization
config              - contains naavik config files
docs                - you are here :). Contains the documentation for the project
internal/           - contains the source code of the project which is not meant to be used by other projects
├── bootstrap       - contains the bootstrap code for the project
    ├── controllers.go  - contains the controller initialization code
    ├── init.go         - contains the naavik initialization code
    ├── profiler.go     - contains the profiler initialization code
    ├── server.go       - contains the server initialization code
├── cache           - contains the global cache implementation for various resources
├── controller      - contains the controller implementation for various resources
├── executor        - contains the executor implementations which allows processing events in desired manner.
├── fake            - contains the fake implementations of various interfaces for testing
├── handlers        - contains the handlers for various resources which does the actual processing of the events
├── lease_checker   - contains the lease checker implementation, used for discovering the leader who will process the events
├── server          - contains the server implementation which is used to implement the apis
├── types           - contains the types for various resources
└── utils           - contains the utility functions that are used across the project
manifests           - contains the kubernetes manifests. Example samples, crds and other yamls for the project
pkg/                - contains the source code of the project which can be used by other projects
├── clients         - contains the wrapper clients for various dependencies
├── logger          - contains the logger implementation as a wrapper over multiple loggers
├── types           - contains the types for various resources
└── utils           - contains the utility functions that are used across the project
```