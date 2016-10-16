package main

func main() {
	conf := GetConfig()
	if conf.Print {
		conf.PrintConf()
	} else {
		SaveData(Portal(conf), conf.Output)
	}
}
