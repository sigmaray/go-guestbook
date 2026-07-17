killair:
	kill -9 `lsof -t -i:8084`; air s

.PHONY: killair
