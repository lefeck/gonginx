package config

// Config represents a complete nginx configuration file.
type Config struct {
	*Block
	FilePath string
}

// Global wrappers provide extension points for custom directive handling.
var (
	BlockWrappers     = map[string]func(*Directive) (IDirective, error){}
	DirectiveWrappers = map[string]func(*Directive) (IDirective, error){}
	IncludeWrappers   = map[string]func(*Directive) (IDirective, error){}
)

// 1. 加载include文件
// 2. 返回到 Config
// 3. 解析Config
// 代码实现
// NewConfig creates a new config from a block
// include conf.d/*.conf;
// include conf.d/app.conf;
//func WholeDir(file string) ([]string, error) {
//	// 1. 找inclue指令后面的值, 如果目录下是一个文件, 那么就直接返回这个文件的内容, 否则, 遍历查找目录下的所有*.conf文件
//
//	// 2. 然后, 解析每一个配置文件
//
//}

//TODO(tufan): move that part inti dumper package
//SaveToFile save config to a file
//TODO: add custom file / folder path support
//func (c *Config) SaveToFile(style *dumper.Style) error {
//	//wrilte file
//	dirPath := filepath.Dir(c.FilePath)
//	if _, err := os.Stat(dirPath); err != nil {
//		err := os.MkdirAll(dirPath, os.ModePerm)
//		if err != nil {
//			return err //TODO: do we reallt need to find a way to test dir creating error?
//		}
//	}
//
//	//write main file
//	err := ioutil.WriteFile(c.FilePath, c.ToByteArray(style), 0644)
//	if err != nil {
//		return err //TODO: do we need to find a way to test writing error?
//	}
//
//	//write sub files (inlude /file/path)
//	for _, directive := range c.Block.Directives {
//		if fs, ok := (interface{}(directive)).(FileDirective); ok {
//			err := fs.SaveToFile(style)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}

// FindDirectives find directives from whole config block
func (c *Config) FindDirectives(directiveName string) []IDirective {
	return c.Block.FindDirectives(directiveName)
}

// FindUpstreams find directives from whole config block
func (c *Config) FindUpstreams() []*Upstream {
	var upstreams []*Upstream
	directives := c.Block.FindDirectives("upstream")
	for _, directive := range directives {
		//	up, _ := NewUpstream(directive)
		upstreams = append(upstreams, directive.(*Upstream))
	}
	return upstreams
}

// FindServersByName finds servers by server_name directive
func (c *Config) FindServersByName(name string) []*Server {
	var servers []*Server
	serverDirectives := c.Block.FindDirectives("server")

	for _, directive := range serverDirectives {
		if server, ok := directive.(*Server); ok {
			// Check server_name directives within this server block
			serverNameDirectives := server.FindDirectives("server_name")
			for _, serverNameDir := range serverNameDirectives {
				params := serverNameDir.GetParameters()
				for _, param := range params {
					if param.GetValue() == name {
						servers = append(servers, server)
						break
					}
				}
			}
		}
	}
	return servers
}

// FindUpstreamByName finds upstream by name
func (c *Config) FindUpstreamByName(name string) *Upstream {
	upstreams := c.FindUpstreams()
	for _, upstream := range upstreams {
		if upstream.UpstreamName == name {
			return upstream
		}
	}
	return nil
}

// FindLocationsByPattern finds locations by pattern (exact match or regex)
func (c *Config) FindLocationsByPattern(pattern string) []*Location {
	var locations []*Location
	c.findLocationsRecursive(c.Block, pattern, &locations)
	return locations
}

// findLocationsRecursive recursively finds locations in blocks
func (c *Config) findLocationsRecursive(block IBlock, pattern string, locations *[]*Location) {
	if block == nil {
		return
	}

	directives := block.GetDirectives()
	for _, directive := range directives {
		if directive.GetName() == "location" {
			if location, ok := directive.(*Location); ok {
				// Check if pattern matches location's match string
				// For locations with modifiers, we need to check the full pattern
				var fullPattern string
				if location.Modifier != "" {
					fullPattern = location.Modifier + " " + location.Match
				} else {
					fullPattern = location.Match
				}

				if fullPattern == pattern || location.Match == pattern {
					*locations = append(*locations, location)
				}
			}
		}

		// Recursively search in nested blocks
		if directive.GetBlock() != nil {
			c.findLocationsRecursive(directive.GetBlock(), pattern, locations)
		}
	}
}

// GetAllSSLCertificates gets all SSL certificate paths from the configuration
func (c *Config) GetAllSSLCertificates() []string {
	var certificates []string
	c.findSSLCertificatesRecursive(c.Block, &certificates)
	return certificates
}

// findSSLCertificatesRecursive recursively finds SSL certificates in blocks
func (c *Config) findSSLCertificatesRecursive(block IBlock, certificates *[]string) {
	if block == nil {
		return
	}

	directives := block.GetDirectives()
	for _, directive := range directives {
		if directive.GetName() == "ssl_certificate" {
			params := directive.GetParameters()
			if len(params) > 0 {
				*certificates = append(*certificates, params[0].GetValue())
			}
		}

		// Recursively search in nested blocks
		if directive.GetBlock() != nil {
			c.findSSLCertificatesRecursive(directive.GetBlock(), certificates)
		}
	}
}

// GetAllUpstreamServers gets all upstream servers from all upstream blocks
func (c *Config) GetAllUpstreamServers() []*UpstreamServer {
	var allServers []*UpstreamServer
	upstreams := c.FindUpstreams()

	for _, upstream := range upstreams {
		allServers = append(allServers, upstream.UpstreamServers...)
	}

	return allServers
}

// FindMaps finds all map blocks in the configuration
func (c *Config) FindMaps() []*Map {
	var maps []*Map
	directives := c.Block.FindDirectives("map")
	for _, directive := range directives {
		if mapBlock, ok := directive.(*Map); ok {
			maps = append(maps, mapBlock)
		}
	}
	return maps
}

// FindMapByVariables finds a map block by its source and target variables
func (c *Config) FindMapByVariables(sourceVar, targetVar string) *Map {
	maps := c.FindMaps()
	for _, mapBlock := range maps {
		if mapBlock.Variable == sourceVar && mapBlock.MappedVariable == targetVar {
			return mapBlock
		}
	}
	return nil
}

// FindGeos finds all geo blocks in the configuration
func (c *Config) FindGeos() []*Geo {
	var geos []*Geo
	directives := c.Block.FindDirectives("geo")
	for _, directive := range directives {
		if geoBlock, ok := directive.(*Geo); ok {
			geos = append(geos, geoBlock)
		}
	}
	return geos
}

// FindGeoByVariable finds a geo block by its target variable
func (c *Config) FindGeoByVariable(targetVar string) *Geo {
	geos := c.FindGeos()
	for _, geoBlock := range geos {
		if geoBlock.Variable == targetVar {
			return geoBlock
		}
	}
	return nil
}

// FindGeoByVariables finds a geo block by its source and target variables
func (c *Config) FindGeoByVariables(sourceVar, targetVar string) *Geo {
	geos := c.FindGeos()
	for _, geoBlock := range geos {
		if geoBlock.SourceAddress == sourceVar && geoBlock.Variable == targetVar {
			return geoBlock
		}
	}
	return nil
}

// FindSplitClients finds all split_clients blocks in the configuration
func (c *Config) FindSplitClients() []*SplitClients {
	var splitClients []*SplitClients
	directives := c.Block.FindDirectives("split_clients")
	for _, directive := range directives {
		if splitClientsBlock, ok := directive.(*SplitClients); ok {
			splitClients = append(splitClients, splitClientsBlock)
		}
	}
	return splitClients
}

// FindSplitClientsByVariable finds a split_clients block by its target variable
func (c *Config) FindSplitClientsByVariable(targetVar string) *SplitClients {
	splitClients := c.FindSplitClients()
	for _, splitClientsBlock := range splitClients {
		if splitClientsBlock.MappedVariable == targetVar {
			return splitClientsBlock
		}
	}
	return nil
}

// FindSplitClientsByVariables finds a split_clients block by its source and target variables
func (c *Config) FindSplitClientsByVariables(sourceVar, targetVar string) *SplitClients {
	splitClients := c.FindSplitClients()
	for _, splitClientsBlock := range splitClients {
		if splitClientsBlock.Variable == sourceVar && splitClientsBlock.MappedVariable == targetVar {
			return splitClientsBlock
		}
	}
	return nil
}

// FindLimitReqZones finds all limit_req_zone directives in the configuration
func (c *Config) FindLimitReqZones() []*LimitReqZone {
	var limitReqZones []*LimitReqZone
	directives := c.Block.FindDirectives("limit_req_zone")
	for _, directive := range directives {
		if limitReqZone, ok := directive.(*LimitReqZone); ok {
			limitReqZones = append(limitReqZones, limitReqZone)
		}
	}
	return limitReqZones
}

// FindLimitReqZoneByName finds a limit_req_zone directive by its zone name
func (c *Config) FindLimitReqZoneByName(zoneName string) *LimitReqZone {
	limitReqZones := c.FindLimitReqZones()
	for _, zone := range limitReqZones {
		if zone.ZoneName == zoneName {
			return zone
		}
	}
	return nil
}

// FindLimitReqZonesByKey finds limit_req_zone directives by their key variable
func (c *Config) FindLimitReqZonesByKey(key string) []*LimitReqZone {
	var zones []*LimitReqZone
	limitReqZones := c.FindLimitReqZones()
	for _, zone := range limitReqZones {
		if zone.Key == key {
			zones = append(zones, zone)
		}
	}
	return zones
}

// FindLimitConnZones finds all limit_conn_zone directives in the configuration
func (c *Config) FindLimitConnZones() []*LimitConnZone {
	var limitConnZones []*LimitConnZone
	directives := c.Block.FindDirectives("limit_conn_zone")
	for _, directive := range directives {
		if limitConnZone, ok := directive.(*LimitConnZone); ok {
			limitConnZones = append(limitConnZones, limitConnZone)
		}
	}
	return limitConnZones
}

// FindLimitConnZoneByName finds a limit_conn_zone directive by its zone name
func (c *Config) FindLimitConnZoneByName(zoneName string) *LimitConnZone {
	limitConnZones := c.FindLimitConnZones()
	for _, zone := range limitConnZones {
		if zone.ZoneName == zoneName {
			return zone
		}
	}
	return nil
}

// FindLimitConnZonesByKey finds limit_conn_zone directives by their key variable
func (c *Config) FindLimitConnZonesByKey(key string) []*LimitConnZone {
	var zones []*LimitConnZone
	limitConnZones := c.FindLimitConnZones()
	for _, zone := range limitConnZones {
		if zone.Key == key {
			zones = append(zones, zone)
		}
	}
	return zones
}

// FindProxyCachePaths finds all proxy_cache_path directives in the configuration
func (c *Config) FindProxyCachePaths() []*ProxyCachePath {
	var proxyCachePaths []*ProxyCachePath
	directives := c.Block.FindDirectives("proxy_cache_path")
	for _, directive := range directives {
		if proxyCachePath, ok := directive.(*ProxyCachePath); ok {
			proxyCachePaths = append(proxyCachePaths, proxyCachePath)
		}
	}
	return proxyCachePaths
}

// FindProxyCachePathByZone finds a proxy_cache_path directive by its keys_zone name
func (c *Config) FindProxyCachePathByZone(zoneName string) *ProxyCachePath {
	proxyCachePaths := c.FindProxyCachePaths()
	for _, cachePath := range proxyCachePaths {
		if cachePath.KeysZoneName == zoneName {
			return cachePath
		}
	}
	return nil
}

// FindProxyCachePathsByPath finds proxy_cache_path directives by their cache path
func (c *Config) FindProxyCachePathsByPath(path string) []*ProxyCachePath {
	var cachePaths []*ProxyCachePath
	proxyCachePaths := c.FindProxyCachePaths()
	for _, cachePath := range proxyCachePaths {
		if cachePath.Path == path {
			cachePaths = append(cachePaths, cachePath)
		}
	}
	return cachePaths
}

// FindStreams finds all stream blocks in the configuration
func (c *Config) FindStreams() []*Stream {
	var streams []*Stream
	directives := c.Block.FindDirectives("stream")
	for _, directive := range directives {
		if stream, ok := directive.(*Stream); ok {
			streams = append(streams, stream)
		}
	}
	return streams
}

// FindStreamUpstreams finds all upstream blocks within stream contexts
func (c *Config) FindStreamUpstreams() []*StreamUpstream {
	var allUpstreams []*StreamUpstream
	streams := c.FindStreams()

	for _, stream := range streams {
		upstreams := stream.FindUpstreams()
		allUpstreams = append(allUpstreams, upstreams...)
	}

	return allUpstreams
}

// FindStreamUpstreamByName finds a stream upstream by name
func (c *Config) FindStreamUpstreamByName(name string) *StreamUpstream {
	upstreams := c.FindStreamUpstreams()
	for _, upstream := range upstreams {
		if upstream.UpstreamName == name {
			return upstream
		}
	}
	return nil
}

// FindStreamServers finds all server blocks within stream contexts
func (c *Config) FindStreamServers() []*StreamServer {
	var allServers []*StreamServer
	streams := c.FindStreams()

	for _, stream := range streams {
		servers := stream.FindServers()
		allServers = append(allServers, servers...)
	}

	return allServers
}

// FindStreamServersByListen finds stream servers by their listen directive
func (c *Config) FindStreamServersByListen(listen string) []*StreamServer {
	var servers []*StreamServer
	allServers := c.FindStreamServers()

	for _, server := range allServers {
		listenPorts := server.GetListenPorts()
		for _, port := range listenPorts {
			if port == listen {
				servers = append(servers, server)
				break
			}
		}
	}

	return servers
}

// GetAllStreamUpstreamServers gets all upstream servers from all stream upstream blocks
func (c *Config) GetAllStreamUpstreamServers() []*StreamUpstreamServer {
	var allServers []*StreamUpstreamServer
	upstreams := c.FindStreamUpstreams()

	for _, upstream := range upstreams {
		allServers = append(allServers, upstream.Servers...)
	}

	return allServers
}

func init() {
	BlockWrappers["http"] = func(directive *Directive) (IDirective, error) {
		return NewHTTP(directive)
	}
	BlockWrappers["location"] = func(directive *Directive) (IDirective, error) {
		return NewLocation(directive)
	}
	BlockWrappers["_by_lua_block"] = func(directive *Directive) (IDirective, error) {
		return NewLuaBlock(directive)
	}
	BlockWrappers["server"] = func(directive *Directive) (IDirective, error) {
		return NewServer(directive)
	}
	BlockWrappers["upstream"] = func(directive *Directive) (IDirective, error) {
		return NewUpstream(directive)
	}
	BlockWrappers["map"] = func(directive *Directive) (IDirective, error) {
		return NewMap(directive)
	}
	BlockWrappers["geo"] = func(directive *Directive) (IDirective, error) {
		return NewGeo(directive)
	}
	BlockWrappers["split_clients"] = func(directive *Directive) (IDirective, error) {
		return NewSplitClients(directive)
	}
	BlockWrappers["stream"] = func(directive *Directive) (IDirective, error) {
		return NewStream(directive)
	}
	BlockWrappers["stream_upstream"] = func(directive *Directive) (IDirective, error) {
		return NewStreamUpstream(directive)
	}
	BlockWrappers["stream_server"] = func(directive *Directive) (IDirective, error) {
		return NewStreamServer(directive)
	}

	DirectiveWrappers["server"] = func(directive *Directive) (IDirective, error) {
		return NewUpstreamServer(directive)
	}
	DirectiveWrappers["stream_upstream_server"] = func(directive *Directive) (IDirective, error) {
		return NewStreamUpstreamServer(directive)
	}
	DirectiveWrappers["limit_req_zone"] = func(directive *Directive) (IDirective, error) {
		return NewLimitReqZone(directive)
	}
	DirectiveWrappers["limit_conn_zone"] = func(directive *Directive) (IDirective, error) {
		return NewLimitConnZone(directive)
	}
	DirectiveWrappers["proxy_cache_path"] = func(directive *Directive) (IDirective, error) {
		return NewProxyCachePath(directive)
	}

	IncludeWrappers["include"] = func(directive *Directive) (IDirective, error) {
		return NewInclude(directive)
	}
}
