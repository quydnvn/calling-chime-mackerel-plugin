## Build File
```go
go build <filename>
```
## Run File
```go
go run <filename>
```
## Config Mackerel Plugin
```shell
[plugin.metrics.chime]
command = "<path>/<filename>"
```

## Source code structure
When using go-mackerel-plugin, the plugin source code consists of the following five parts.

1. The package declaration and import statement
2. The plugin struct definition
3. The graph definition output method GraphDefinition defined in struct
4. The metric acquisition method FetchMetrics defined in struct
5. The metric key prefix acquisition method MetricKeyPrefix defined in struct
6. The main() function definition

**1. The package declaration and import statement**
package main

    import (
        "flag"
        "fmt"
        "strings"
        mp "github.com/mackerelio/go-mackerel-plugin"
        "github.com/mackerelio/go-osstat/uptime"
    )

Declare the package as main. Additionally, it’s typical to import go-mackerel-plugin under the alias mp.

**2. The plugin struct definition
**
```go
type UptimePlugin struct {
Prefix string
}
```
This is the definition of the plugin struct. In this struct, the Prefix field is defined. This is used to determine the beginning of the metric namespace when outputting graph definitions. The standard Prefix for the uptime plugin is uptime and metrics are output with the key uptime.seconds, but this field is used when you want to change the uptime to uptime2 for example.

In the uptime plugin, the Prefix field is not particularly useful. However, for middleware plugins for example, if you run the same middleware multiple times on one host and you want to obtain each metric, you’ll need to separate the metric namespaces. Therefore, defining this field is recommended.

In the uptime plugin, only this Prefix field is defined, but if it’s a general middleware plugin, the fields Port and Host will also be required.

Additionally, this plugin struct must satisfy the interface of mp.PluginWithPrefix . The interface definition is as follows.

```go
type PluginWithPrefix interface {
FetchMetrics() (map[string]float64, error)
GraphDefinition() map[string]mp.Graphs
MetricKeyPrefix() string
}
```
This is the graph definition output method, the metric acquisition method, and the metric prefix acquisition method.

**3. The graph definition output method GraphDefinition defined in struct
**
```go
func (u UptimePlugin) GraphDefinition() map[string]mp.Graphs {
labelPrefix := strings.Title(u.MetricKeyPrefix())
return map[string]mp.Graphs{
"": {
Label: labelPrefix,
Unit:  mp.UnitFloat,
Metrics: []mp.Metrics{
{Name: "seconds", Label: "Seconds"},
},
},
}
}
```
By defining this GraphDefinition, the graph definition JSON and metrics will be output correctly. With the uptime plugin, only one graph definition is returned with the empty string as the key, but many plugins return multiple graph definitions with keys like "runtime", "memory".

When the plugin outputs the list of graph definitions, it attaches the prefix returned by the MetricKeyPrefix() function before each key of the graph definition. For example, if MetricKeyPrefix() returns uptime, it’s output as "" -> "uptime", "runtime" -> "uptime.runtime".

Label is the graph name displayed in Mackerel, Unit is the unit of graph, and "float", "integer", "percentage", "bytes", "bytes / sec", "iops" are able to be specified similar to the graph definition API.

**4. The metric acquisition method FetchMetrics defined in struct
**
```go
func (u UptimePlugin) FetchMetrics() (map[string]float64, error) {
ut, err := uptime.Get()
if err != nil {
return nil, fmt.Errorf("Failed to fetch uptime metrics: %s", err)
}
return map[string]float64{"seconds": ut.Seconds()}, nil
}
```
FetchMetrics() returns values in map[string]float64 format. It returns a map with the uptime value stored with the key as seconds. seconds is the specified Name in the GraphDefinition above.

**5. The metric key prefix acquisition method MetricKeyPrefix defined in struct
**
```go
func (u UptimePlugin) MetricKeyPrefix() string {
if u.Prefix == "" {
u.Prefix = "uptime"
}
return u.Prefix
}
```
Define the method for obtaining the metric prefix. If specified by the plugin user, the specified prefix is returned, if not "uptime" is returned by default.

**6. The main() function definition
**

```go
func main() {
optPrefix := flag.String("metric-key-prefix", "uptime", "Metric key prefix")
optTempfile := flag.String("tempfile", "", "Temp file name")
flag.Parse()

u := UptimePlugin{
Prefix: *optPrefix,
}
plugin := mp.NewMackerelPlugin(u)
plugin.Tempfile = *optTempfile
plugin.Run()
}
```
This is the plugin’s main procedure. It parses command line options, creates plugin objects, and runs the plugin.

The Tempfile, specified in plugin, is a file that holds the last values obtained for calculating the difference values. Please note that this should be specified so that overlaps do not occur with other plugins or when there are multiple configurations with varying parameters within the same plugin.
