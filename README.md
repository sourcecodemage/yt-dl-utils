# yt-dlp utils



## Features 

- Youtube videos - download videos and video metadata.
- Youtube Channels - download channel metadata. Create tvshow.nfo file. Create show.jpg file.
- Youtube Playlist - Remove entries from download playload that are in your library. 

NFO File
- Generates [video file name].nfo file for each video
- Generates tvshow.nfo file for each channel


## Installation

### Download and install yt-dlp

### Download and install yt-dlp dependencies 

Install required utilities  

```bash
    brew install ffmpeg coreutils 
```

Install challenge response utilities  required for downloads to work

```bash
    brew install deno quickjs node 
```

Install recommended utilities

```bash
    brew install ffmpeg mkvtoolnix mediainfo jq coreutils 
```


### Build youtube-util go binary

1. Init the repo and mods
    ```
    Go mod init
    go mod tidy 
    ```
2. Run these commands to build the binary
    ```
    go build ./dist/
    ```

## Releases

To publish a release, update `VERSION` to a new version such as `v0.1.0`, then commit and push that change:

```bash
git add VERSION
git commit -m "Release v0.1.0"
git push
```

Create and push a tag with the exact same value as `VERSION` to start the GitHub Actions release workflow:

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

The workflow builds the release binaries and publishes them with checksums in the project's Releases tab. Use a new version for each release; tags must match `VERSION` exactly and existing tags cannot be reused.

## Useful Links

- [yt-dlp project](https://github.com/yt-dlp/yt-dlp)
- [External JS Scripts Setup Guide](https://github.com/yt-dlp/yt-dlp/wiki/ejs)
- [Plex NFO guide](https://support.plex.tv/articles/using-nfo-metadata-files-with-plex/)