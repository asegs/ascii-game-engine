package main

/**
If recorded is in provided list - record number in list, nothing more.
If not, consider recorded erratas and store in file as JSON/something.
 */

func recordedEquals (r1 * Recorded, r2 * Recorded) bool {
	return r1.Format.Format == r2.Format.Format &&
		r1.BackgroundCode == r2.BackgroundCode &&
		r1.ShownSymbol == r2.ShownSymbol
}

func recordedIndex (r * Recorded, styles [] * Recorded) int {
	for i, recorded := range styles {
		if recordedEquals(r,recorded) {
			return i
		}
	}
	return -1
}

func recordedToString(r * Recorded) {

}

func recordedFromString (record string) * Recorded {

}


func (t * Terminal) saveState (filename string,styles [] * Recorded) error{

}

func (t * Terminal) loadState (filename string, styles [] * Recorded) error {

}
