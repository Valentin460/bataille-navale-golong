package models

type BoardResponse struct {
	Size  int     `json:"size"`
	Cells [][]int `json:"cells"`
}

type BoatsResponse struct {
	RemainingBoats int `json:"remaining_boats"`
}

type HitRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type HitResponse struct {
	Result string `json:"result"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

type HitsResponse struct {
	Hits []HitInfo `json:"hits"`
}

type HitInfo struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Result string `json:"result"`
}

// permet à un joueur de s'enregistrer auprès d'un autre
type RegisterRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}