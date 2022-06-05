package costconfigs

import "github.com/grab/async/engine/sample/service/costconfigs/dummy"

func (c computer) doFetch() dummy.MergedCostConfigs {
	return c.fetcher.Fetch()
}
