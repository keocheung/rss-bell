package dlwebhook

type Data struct {
	URL          string `json:"url"`
	Secret       string `json:"secret"`
	Engine       string `json:"engine"`
	Path         string `json:"path"`
	Name         string `json:"name"`
	ExtraOptions string `json:"extra_options"`
}
