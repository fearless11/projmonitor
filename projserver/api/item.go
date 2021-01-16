package api

import "projmonitor/projserver/conf"

type GetItemResponse struct {
	Message string
	Data    []*conf.DetectedItem
}

func (this *Web) GetItem(hostname string, resp *GetItemResponse) error {
	items, exists := conf.DetectedItemMap.Get(hostname)
	if !exists {
		resp.Message = "No project is assigned to the" + hostname
	}
	resp.Data = items
	return nil
}
