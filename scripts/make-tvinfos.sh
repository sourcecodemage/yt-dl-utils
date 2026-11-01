#!/bin/bash

# Find video info.json file and use it's information to create channel tvinfo.json file
while IFS= read -r line; do
	find "${line}" -type f -name "*.info.json" -print0 -depth 2 -quit |  while IFS= read -r -d '' file; do
		printf "Processing file %s\n" "${file}"
		# json=$(< "${file}")
		channel_id=$(jq -r '.id' "${file}")
		channel_url=$(jq -r '.channel_url' "${file}")
		channel=$(jq -r '.channel' "${file}")
		studio=Youtube
		channel_dir=$(dirname "$(dirname "$file")")
		tvinfo_path=$(printf "%s/tvshow.nfo" "${channel_dir}") 
		printf "<tvshow>\n" > "${tvinfo_path}"
		printf "<id>%s</id>\n" "${channel_id}" >> "${tvinfo_path}"
		printf "<title>%s</title>\n" "${channel}"  >> "${tvinfo_path}"
		printf "<plot>%s</plot>\n" "${channel_url}" >> "${tvinfo_path}"
		printf "<studio>Youtube</studio>\n" >> "${tvinfo_path}"
		printf "</tvshow>\n" >> "${tvinfo_path}"
		printf "TVinfo file %s created.\n\n" "${tvinfo_path}"
	done
done < missing-tvshownfo.txt

