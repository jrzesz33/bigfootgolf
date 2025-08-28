package controllers

type AppConfig struct {
	KeyValues map[string]string
}

func GetAppConfig(key string) string {
	//
	_configs := make(map[string]string)

	//hard code for now
	_configs["ANTHROPIC_API_KEY"] = ""
	return _configs["ANTHROPIC_API_KEY"]
}
