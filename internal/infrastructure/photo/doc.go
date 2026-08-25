// Package photo stores gallery files on the mounted volume and generates thumbnails.
//
// Files never go into SQLite. Names are content-addressed; the filename a guest
// uploaded is metadata only and never a path component. Originals are kept
// byte-for-byte with EXIF intact — this is a private archive, and stripping EXIF
// breaks image orientation. Thumbnails are regenerable, so losing them is not data loss.
package photo
