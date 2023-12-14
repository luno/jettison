package log

// ElasticStringArray is a converted type that stops ElasticSearch from joining
// elements of the string slice with commas
type ElasticStringArray []struct {
	Content []string `json:">"`
}

func MakeElastic(s []string) ElasticStringArray {
	if len(s) == 0 {
		return nil
	}
	return ElasticStringArray{{Content: s}}
}

func (e ElasticStringArray) Content() []string {
	if len(e) == 0 {
		return nil
	}
	return e[0].Content
}
