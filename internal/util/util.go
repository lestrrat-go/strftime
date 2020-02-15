package util

/*
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int getweek(int time_stamp) {
	time_t rawtime = time_stamp;
	char buffer[128] = {0};
	struct tm *timeinfo;
	timeinfo = localtime(&rawtime);
	strftime(buffer, sizeof(buffer), "%Y%U", timeinfo);
	return atoi(buffer);
}
*/
import "C"

func GHIssue12GetWeek(in_time int64) int {
	cw := C.getweek(C.int(in_time))
	return int(cw)
}

