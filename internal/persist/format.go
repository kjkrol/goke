package persist

// Magic identifies a goke save file; the first bytes written by Save and
// checked by Load.
const Magic = "GKSV"

// FormatVersion is the current save-file format version.
const FormatVersion uint32 = 1
