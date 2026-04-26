package main

func applyWindowTitle(title string) {
	debugPrintln("Window changed:", title)

	if matched, rate := matchWindow(title); matched {
		debugPrintf("Matched target window, setting %dk\n", rate)
		setPollingRate(rate)
	} else {
		debugPrintln("Defaulting to 8k")
		set8k()
	}

	setConfig()
}
