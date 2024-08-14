package serv

import (
	"cirno-im/constants"
	"hash/crc32"
	"math/rand"

	"cirno-im"
	"cirno-im/logger"
	"cirno-im/services/gateway/conf"
	"cirno-im/wire/pkt"
)

// RouteSelector RouteSelector
type RouteSelector struct {
	route *conf.Route
}

func NewRouteSelector(configPath string) (*RouteSelector, error) {
	route, err := conf.ReadRoute(configPath)
	if err != nil {
		return nil, err
	}
	return &RouteSelector{
		route: route,
	}, nil
}

// Lookup a server
func (s *RouteSelector) Lookup(header *pkt.Header, srvs []cim.Service) string {
	// 1. 从header中读取Meta信息
	app, _ := pkt.FindMeta(header.Meta, constants.MetaKeyApp)
	account, _ := pkt.FindMeta(header.Meta, constants.MetaKeyAccount)
	if app == nil || account == nil {
		ri := rand.Intn(len(srvs))
		return srvs[ri].ServiceID()
	}
	log := logger.WithFields(logger.Fields{
		"app":     app,
		"account": account,
	})

	// 2. 判断是否命中白名单
	zone, ok := s.route.Whitelist[app.(string)]
	if !ok { // 未命中情况
		var key string
		switch s.route.RouteBy {
		case constants.MetaKeyApp:
			key = app.(string)
		case constants.MetaKeyAccount:
			key = account.(string)
		default:
			key = account.(string)
		}
		// 3. 通过权重计算出zone
		slot := hashcode(key) % len(s.route.Slots)
		i := s.route.Slots[slot]
		zone = s.route.Zones[i].ID
	} else {
		log.Infoln("hit a zone in whitelist", zone)
	}
	// 4. 过滤出当前zone的servers
	zoneSrvs := filterSrvs(srvs, zone)
	if len(zoneSrvs) == 0 {
		noServerFoundErrorTotal.WithLabelValues(zone).Inc()
		log.Warnf("select a random service from all due to no service found in zone %s", zone)
		ri := rand.Intn(len(srvs))
		return srvs[ri].ServiceID()
	}
	// 5. 从zoneSrvs中选中一个服务
	srv := selectSrvs(zoneSrvs, account.(string))
	return srv.ServiceID()
}

func filterSrvs(srvs []cim.Service, zone string) []cim.Service {
	var res = make([]cim.Service, 0, len(srvs))
	for _, srv := range srvs {
		if zone == srv.GetMetadata()["zone"] {
			res = append(res, srv)
		}
	}
	return res
}

func selectSrvs(srvs []cim.Service, account string) cim.Service {
	slots := make([]int, 0, len(srvs)*10)
	for i := range srvs {
		for j := 0; j < 10; j++ {
			slots = append(slots, i)
		}
	}
	slot := hashcode(account) % len(slots)
	return srvs[slots[slot]]
}

func hashcode(key string) int {
	hash32 := crc32.NewIEEE()
	hash32.Write([]byte(key))
	return int(hash32.Sum32())
}
