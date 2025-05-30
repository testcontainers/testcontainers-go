The following options are exposed by the `testcontainers` package.

### Basic Options

- [`WithExposedPorts`](/features/creating_container/#withexposedports) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>
- [`WithEnv`](/features/creating_container/#withenv) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>
- [`WithWaitStrategy`](/features/creating_container/#withwaitstrategy) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>
- [`WithAdditionalWaitStrategy`](/features/creating_container/#withadditionalwaitstrategy) Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>
- [`WithWaitStrategyAndDeadline`](/features/creating_container/#withwaitstrategyanddeadline) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>
- [`WithAdditionalWaitStrategyAndDeadline`](/features/creating_container/#withadditionalwaitstrategyanddeadline) Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>
- [`WithEntrypoint`](/features/creating_container/#withentrypoint) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>
- [`WithEntrypointArgs`](/features/creating_container/#withentrypointargs) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>
- [`WithCmd`](/features/creating_container/#withcmd) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>
- [`WithCmdArgs`](/features/creating_container/#withcmdargs) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>
- [`WithLabels`](/features/creating_container/#withlabels) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

### Lifecycle Options

- [`WithLifecycleHooks`](/features/creating_container/#withlifecyclehooks) Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>
- [`WithAdditionalLifecycleHooks`](/features/creating_container/#withadditionallifecyclehooks) Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>
- [`WithStartupCommand`](/features/creating_container/#withstartupcommand) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.25.0"><span class="tc-version">:material-tag: v0.25.0</span></a>
- [`WithAfterReadyCommand`](/features/creating_container/#withafterreadycommand) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>

### Files & Mounts Options

- [`WithFiles`](/features/creating_container/#withfiles) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>
- [`WithMounts`](/features/creating_container/#withmounts) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>
- [`WithTmpfs`](/features/creating_container/#withtmpfs) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>
- [`WithImageMount`](/features/creating_container/#withimagemount) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

### Build Options

- [`WithDockerfile`](/features/creating_container/#withdockerfile) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>

### Logging Options

- [`WithLogConsumers`](/features/creating_container/#withlogconsumers) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.28.0"><span class="tc-version">:material-tag: v0.28.0</span></a>
- [`WithLogConsumerConfig`](/features/creating_container/#withlogconsumerconfig) Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>
- [`WithLogger`](/features/creating_container/#withlogger) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.29.0"><span class="tc-version">:material-tag: v0.29.0</span></a>

### Image Options

- [`WithAlwaysPull`](/features/creating_container/#withalwayspull) Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>
- [`WithImageSubstitutors`](/features/creating_container/#withimagesubstitutors) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.26.0"><span class="tc-version">:material-tag: v0.26.0</span></a>
- [`WithImagePlatform`](/features/creating_container/#withimageplatform) Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

### Networking Options

- [`WithNetwork`](/features/creating_container/#withnetwork) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.27.0"><span class="tc-version">:material-tag: v0.27.0</span></a>
- [`WithNetworkByName`](/features/creating_container/#withnetworkbyname) Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>
- [`WithBridgeNetwork`](/features/creating_container/#withbridgenetwork) Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>
- [`WithNewNetwork`](/features/creating_container/#withnewnetwork) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.27.0"><span class="tc-version">:material-tag: v0.27.0</span></a>

### Advanced Options

- [`WithHostPortAccess`](/features/creating_container/#withhostportaccess) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.31.0"><span class="tc-version">:material-tag: v0.31.0</span></a>
- [`WithConfigModifier`](/features/creating_container/#withconfigmodifier) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>
- [`WithHostConfigModifier`](/features/creating_container/#withhostconfigmodifier) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>
- [`WithEndpointSettingsModifier`](/features/creating_container/#withendpointsettingsmodifier) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>
- [`CustomizeRequest`](/features/creating_container/#customizerequest) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.20.0"><span class="tc-version">:material-tag: v0.20.0</span></a>
- [`WithName`](/features/creating_container/#withname) Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>
- [`WithNoStart`](/features/creating_container/#withnostart) Not available until the next release <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>

### Experimental Options

- [`WithReuseByName`](/features/creating_container/#withreusebyname) Since <a href="https://github.com/testcontainers/testcontainers-go/releases/tag/v0.37.0"><span class="tc-version">:material-tag: v0.37.0</span></a>
