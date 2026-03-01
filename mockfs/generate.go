package mockfs

//go:generate go tool mockgen -destination=mock_fs.go -package=mockfs github.com/unstoppablemango/ihfs ChmodFS,ChownFS,ChtimesFS,CopyFS,CreateFS,CreateTempFS,GlobFS,LinkerFS,MkdirFS,MkdirAllFS,MkdirTempFS,OpenFileFS,OsFS,ReadDirFS,ReadDirNamesFS,ReadFileFS,ReadLinkFS,RemoveFS,RemoveAllFS,RenameFS,StatFS,SubFS,SymlinkFS,TempFileFS,WriteFileFS
//go:generate go tool mockgen -destination=mock_file.go -package=mockfs github.com/unstoppablemango/ihfs DirNameReader,DirReader,File,Operation,ReaderAt,ReadDirFile,Seeker,StringWriter,Syncer,Truncater,Writer,WriterAt
