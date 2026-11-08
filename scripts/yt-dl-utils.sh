#!/bin/bash

# printf "%s START %s script starting.\n" "$(date +"%Y%m%d_%H%M%S")" "${0}"
set -eEo pipefail -o functrace

declare -a REPLACE_IN_METADATA_ARGS=(
        '--replace-in-metadata channel,title "/" "⧸"'
        '--replace-in-metadata channel,title "\?" ""'
        '--replace-in-metadata channel,title ":" "-"'
        '--replace-in-metadata channel,title "\*" ""'
)

        # '--replace-in-metadata "channel,title" "ä" "ae"'
        # '--replace-in-metadata "channel,title" "ß" "ss"'
        # '--replace-in-metadata "channel,title" "ë" "e"'
        # '--replace-in-metadata "channel,title" "ö" "oe"'
        # '--replace-in-metadata "channel,title" "İ" "I"'
        # '--replace-in-metadata "channel,title" "ş" "s"'
        # '--replace-in-metadata "channel,title" "ü" "ue"'

function cleanup_temp_dir {
    if [[ -n "${FIND_FILE:-}" && -f "${FIND_FILE}" ]]; then 
        # first delete .DS_Store files
        "${FIND_FILE}" "${TEMP_DIR}" -type f -mindepth 1 -name ".DS_Store" -delete
        # Delete empty subdirectories
        "${FIND_FILE}" "${TEMP_DIR}" -type d -mindepth 1 -delete -empty
        if [ ! -z "$("${FIND_FILE}" "${TEMP_DIR}" -mindepth 1)" ]; then
            printf "%s contains directories.\n" "${TEMP_DIR}"
            return 1
        fi
        printf "%s has no subdirectories. Directory clean." "${TEMP_DIR}"
    fi
}

error_handler() {
    local exit_code=$1
    local line_number=$2
    local command="$BASH_COMMAND"
    
    echo "ERROR: Command '$command' failed on line $line_number with exit code $exit_code." >&2
    exit "$exit_code"
}

# 4. Bind the handler to the ERR signal
trap 'error_handler $? $LINENO' ERR

function check_required_variables {
    printf "Checking that all required environment variables exist.\n"
    # List required environment variables
    required_vars=(
        CHANNEL_LIST_URLS
        CHANNEL_UNAVAILABLE_FILE
        DOWNLOAD_ARCHIVE_FILE
        FIND_FILE
        HOME_DIR
        INVENTORY_FILE
        LISTS_DIR
        VIDEO_BLACKLIST_FILE
        TEMP_DIR
        VIDEO_UNAVAILABLE_FILE
        YTDLP_FILE
    )

    missing=()

    for var in "${required_vars[@]}"; do
        # Check if variable is unset or empty
        if [[ -z "${!var:-}" ]]; then
            missing+=("$var")
        fi
    done

    if (( ${#missing[@]} > 0 )); then
        printf "ERROR: The following required environment variables are missing or empty:"
        printf '  - %s\n' "${missing[@]}"
        exit 1
    fi
    printf "All required environment variables exist!\n\n"
}


function check_required_directories {
    printf "Checking that required directories exist.\n"
    # check for required directories 
    required_dirs=(
        "${HOME_DIR}"
        "${LISTS_DIR}"
        "${TEMP_DIR}"
    )

    missing=()

    for var in "${required_dirs[@]}"; do
        # Check if director exists and is a directory
        if [[ ! -d "${var}" ]]; then
            missing+=("$var")
        fi
    done

    if (( ${#missing[@]} > 0 )); then
        echo "ERROR: The following required directories are missing:"
        printf '  - %s\n' "${missing[@]}"
        exit 1
    fi
    printf "All required directories exist!!\n\n"
}

function check_required_files {
    printf "Checking that all required files exist.\n"
    # check for required files 
    required_files=(
        "${CHANNEL_UNAVAILABLE_FILE}"
        "${DOWNLOAD_ARCHIVE_FILE}"
        "${FIND_FILE}"
        "${INVENTORY_FILE}"
        "${YTDLP_FILE}"
    )

    missing=()

    for var in "${required_files[@]}"; do
        # Check if file exists and is a regular file
        if [[ ! -f "${var}" ]]; then
            missing+=("$var")
        fi
    done

    if (( ${#missing[@]} > 0 )); then
        echo "ERROR: The following required files are missing:"
        printf '  - %s\n' "${missing[@]}"
        exit 1
    fi
    printf "All required files exist!\n\n"
}

# updates yt-dlp to the latest nightly branch commit if it has not already been updated today
function update_ytdlp {
    SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]:-$0}"; )" &> /dev/null && pwd 2> /dev/null; )"
    printf -v STAMP_FILE "%s/../conf/YTDLP_UPDATE_STAMP.txt" "${SCRIPT_DIR}"
    TODAY=$(date +%Y-%m-%d)

    # Check if the file exists and matches today's date
    if [[ -f "$STAMP_FILE" ]] && [[ "$(cat "$STAMP_FILE")" == "$TODAY" ]]; then
        echo "Command already ran today ($TODAY). Skipping."
        exit 0
    fi

    PRE_VERSION=$("${YTDLP_FILE}" --version)
    printf "Updating %s with VERSION %s to latest nightly branch commit.\n" "${YTDLP_FILE}" "${PRE_VERSION}"
    "${YTDLP_FILE}" --update-to nightly
    POST_VERSION=$("${YTDLP_FILE}" --version)
    printf "Updated %s from version %s to %s which is the latest nightly branch commit!\n\n" "${YTDLP_FILE}" "${PRE_VERSION}" "${POST_VERSION}"

    # Record today's date only if the command succeeded
    echo "$TODAY" > "$STAMP_FILE"
}


function validate_batch_file {

    # create array of every line in the batc file that is not a comment
    batch_file_lines=$(grep -v "^#" "${BATCH_FILE}")
    # remove any lines that are valid youtube video urls
    batch_file_lines=$(echo "${batch_file_lines}" | grep -vE "https://www.youtube.com/watch\?v=.{11}" || true)
    # remove any lines that are valid youtube playlist urls
    batch_file_lines=$(echo "${batch_file_lines}" | grep -vE "https://www.youtube.com/playlist\?list=PL[^&]+" || true)
    # remove any lines that are valid youtube channel urls
    batch_file_lines=$(echo "${batch_file_lines}" | grep -vE "https://www.youtube.com/(@.+|c/.+|user/.+)/videos" || true)
    if [[ "${batch_file_lines}" ]];then
        printf "ERROR: %s Malformed URL(s) found in batch file %s. Exiting...\n" "${batch_file_lines[@]}" "${BATCH_FILE}"
        printf "Malformed URL(s):\n'%s'\n" "${batch_file_lines}"
        return 1
    else
        printf "All URLs in batch file %s are valid!\n" "${BATCH_FILE}"
        return 0
    fi



}

# sets the batch file to used for downloading videos and metadata.
function set_batch_file  {
    printf "Setting batch file...\n"
    # set the batch file to be used in this run
    if [[ -n "${WANTED_HOME_DIR:-}" ]]; then
        export BATCH_FILE="${WANTED_BATCH_LIST_URLS}"
    else
        export BATCH_FILE="${BATCH_LIST_URLS}"
    fi
    printf "Batch file set to %s!\n\n" "${BATCH_FILE}"
    return 0
}


# prints help
function print_help {
    printf 'Usage: %s -a ACTION -e ENV_FILE\n\n' "$(basename "$0")"
    printf 'Required options:\n'
    printf '  -a ACTION    Action to perform:\n'
    printf '                 download-videos\n'
    printf '                 download-video-metadata\n'
    printf '                 download-channel-metadata\n'
    printf '                 generate-channel-urls\n'
    printf '  -e ENV_FILE  Path to the environment configuration file\n\n'
    printf 'Optional options:\n'
    printf '  -h           Show this help message\n'
}

function initialize_log_file {
    printf "Initializing log file variables...\n"
    # Set log file for video download
    printf -v DOWNLOAD_VIDEOS_LOG_FILE_PATH "%s/%s.download_videos.log" "${LOG_DIR}" "${0##*/}"
    # Set log file for video metadata download
    printf -v DOWNLOAD_VIDEO_METADATA_LOG_FILE_PATH "%s/%s.download_video_metadata.log" "${LOG_DIR}" "${0##*/}"
    # Set log file for channel metadata download
    printf -v DOWNLOAD_CHANNEL_METADATA_LOG_FILE_PATH "%s/%s.download_channel_metadata.log" "${LOG_DIR}" "${0##*/}"
    printf "Log file variables initialized!\n\n"
}

function parse_arguments {
    printf "Parsing command line arguments...\n"
    # Process command line args
    while getopts 'a:e:h' opt; do
        case "$opt" in
            a)
                echo "Processing option 'action'"
                ACTION="${OPTARG}"
                ;;
            e)
                echo "Processing option 'env_file'"
                ENV_FILE="${OPTARG}"
                ;;
            h)
                print_help
                exit 0
                ;;
            \?)
                print_help >&2
                exit 1
                ;;
        esac
    done
    shift $((OPTIND-1))

    # check mandatory arguments 
    # Track if any requirement fails
    missing_count=0

    # define required
    required_arguments=(
        ACTION
        ENV_FILE
    )

    # Loop through the array of variable names
    for var in "${required_arguments[@]}"; do
        # ${!var} uses indirect expansion to check the value of the variable named by $var
        if [[ -z "${!var}" ]]; then
            printf "Error: Required variable '%s' is not set or is empty." "$var"
            ((missing_count++))
        fi
    done

    # Exit if any variable was missing
    if (( missing_count > 0 )); then
        printf "Initialization failed. Missing %s required argument(s).\n\n" "${missing_count}"
        print_help
        exit 1
    fi

    printf "All required arguments are set successfully!\n\n"
    
}

function load_env {
    printf "Sourcing environment file...\n"
    # Source env file. Throw error if it does not exist.
    if [ -f "${ENV_FILE}" ]; then
        source "${ENV_FILE}"
    else
        printf "Env file %s not found. Exiting..." "${ENV_FILE}"
        exit 1
    fi
    printf "Environment file %s sourced!\n\n" "${ENV_FILE}"

}

function set_cookie_cmd {

    printf "Setting the cookie command...\n"
    COOKIE_ARGS=()
    if [[ -n "${COOKIES_FROM_FILE:-}" ]];then 
        COOKIE_ARGS=(--cookies "${COOKIES_FROM_FILE}")
    elif [[ -n "${COOKIES_FROM_BROWSER:-}" ]];then 
        COOKIE_ARGS=(--cookies-from-browser "${COOKIES_FROM_BROWSER}")
    fi
    printf "Cookie command set to %s\n\n" "${COOKIE_ARGS[*]}"


}

# generates a list of channel urls from the info.json files in the library
function generate_channel_urls_file {
    # printf "Generating fresh channel urls file at %s\n." "${CHANNEL_LIST_URLS}"
    # load the unavailable channel ids into a variable for quick lookup
    local unavailable_ids
    unavailable_ids=$'\n'$(<"${CHANNEL_UNAVAILABLE_FILE}")$'\n'
    local channel_url_array=()

    while IFS= read -r info_json_file; do
        if [[ -f "${info_json_file}" ]]; then
            if jq -er '.channel_url' "${info_json_file}" &>/dev/null && jq -er '.channel_id' "${info_json_file}" &>/dev/null; then
                channel_url=$(jq -r '.channel_url' "${info_json_file}")
                channel_id=$(jq -r '.channel_id' "${info_json_file}")
                # check if channel_id is in the unavailable_ids list. skip if it is
                case "${unavailable_ids}" in
                    *$'\n'"${channel_id}"$'\n'*)
                        # printf "Skipping unavailable channel %s.\n" "${channel_id}"
                        continue
                        ;;
                esac

                if [[ -n "${channel_url}" && "${channel_url}" == *"channel/UC"* ]]; then
                    # printf "Validated UC channel_url %s in %s\n" "${channel_url}" "${info_json_file}"
                    # echo "${channel_url}" >> "${CHANNEL_LIST_URLS}"
                    # check if channel_url is already in the array
                    local exists=false
                    for element in "${channel_url_array[@]}"; do
                        if [[ "$element" == "$channel_url" ]]; then
                            exists=true
                            break
                        fi
                    done
                    if [[ "${exists}" == false ]]; then
                        channel_url_array+=("${channel_url}")
                        # printf "Added channel_url %s to array\n" "${channel_url}"
                    fi
                    
                    # CHANNEL_URL_FOUND=true
                    # exit file loop
                    continue
                    # printf "Channel URL found for %s in season dir %s and added to %s\n" "${channel_dir}" "${season_dir}" "${CHANNEL_LIST_URLS}"
                fi
            fi
        fi  
    done < <("${FIND_FILE}" "${HOME_DIR}" -type f -mindepth 2 -maxdepth 2 -name "*.info.json")
    # Write channel_url_array to CHANNEL_LIST_URLS file, sorted and unique
    printf '%s\n' "${channel_url_array[@]}" | sort -u > "${CHANNEL_LIST_URLS}"
    printf "Generated channel urls file at %s with %s entries\n\n!" "${CHANNEL_LIST_URLS}" "$(cat "${CHANNEL_LIST_URLS}" | wc -l)"
}

# generates download archive file from current video inventory
function generate_download_archive {
    # Create inventory list
    printf "Generating Download Archive for library %s\n" "${HOME_DIR}"
    # cat "${INVENTORY_FILE}" "${BLACKLIST_FILE}" "${PRIVATED_FILE}" "${UNAVAILABLE_FILE}" > "${DOWNLOAD_ARCHIVE_FILE}"

    # join columns 1 and 2 ( youtube and video id , minuse the error message column ) of files into the download archive file to skip them    
    cat "${INVENTORY_FILE}" "${VIDEO_BLACKLIST_FILE}" "${VIDEO_UNAVAILABLE_FILE}" | gcut -d' ' --complement -f3-  > "${DOWNLOAD_ARCHIVE_FILE}"
    # Sanity check to ensure that the download archive file was created successfully
    if [[ ! -f "${DOWNLOAD_ARCHIVE_FILE}" || ! -s "${DOWNLOAD_ARCHIVE_FILE}" ]]; then
        printf "ERROR: Download archive file %s was not created successfully. Exiting...\n" "${DOWNLOAD_ARCHIVE_FILE}"
        exit 1
    fi
    
    # download archive is complete
    printf "Download archive file for library %s created at %s\n\n" "${HOME_DIR}" "${DOWNLOAD_ARCHIVE_FILE}"

}

function generate_inventory_file {
    printf "Generating inventory file for library %s\n" "${HOME_DIR}"
    "${FIND_FILE}" "${HOME_DIR}" -type f \( -name "*.mp4" -o -name "*.mkv" -o -name "*.webm" \) | sed -r 's/^.*\[(.*)\]\.(mp4|mkv|webm)$/\1/' | while read -r i;do echo "youtube ${i}";done > "${INVENTORY_FILE}"
    printf "Inventory file for library %s created at %s\n\n" "${HOME_DIR}" "${INVENTORY_FILE}"
}   

function generate_wanted_list {
    printf "Generating Wanted list.\n"
    if [[ -n "${WANTED_HOME_DIR:-}" ]]; then
        # create wanted list
        printf "Generating Wanted list for wanted library %s\n" "${WANTED_HOME_DIR}"
        "${FIND_FILE}" "${WANTED_HOME_DIR}" -type f \( -name "*.mp4" -o -name "*.mkv" -o -name "*.webm" \) | sed -r 's/^.*\[(.*)\]\.(mp4|mkv|webm)$/\1/' | while read -r i;do echo "youtube ${i}";done > "${WANTED_LIST_FILE}"
        # Create download batch list that equals wanted list - inventory list - blacklist
        printf "Generating wanted batch list file  %s\n" "${WANTED_BATCH_LIST_FILE}"
        grep -v -f "${BLACKLIST_FILE}" "${WANTED_LIST_FILE}" > "${WANTED_BATCH_LIST_FILE}" || true
        # validate that the wanted batch list file was created 
        if [[ ! -f "${WANTED_BATCH_LIST_FILE}" || ! -s "${WANTED_BATCH_LIST_FILE}" ]]; then
            printf "ERROR: Wanted batch list file %s was not created successfully. Exiting...\n" "${WANTED_BATCH_LIST_FILE}"
            exit 1
        fi

        # Convert batch list file to youtbe batch url list format for yt-dlp to consume
        printf "Generating wanted batch list urls file  %s\n" "${WANTED_BATCH_LIST_URLS}"
        sed -r 's/^youtube (.*)$/https:\/\/www.youtube.com\/watch?v=\1/' "${WANTED_BATCH_LIST_FILE}" > "${WANTED_BATCH_LIST_URLS}"
    fi 

    printf "Wanted list generated!\n\n"

}

# downloads youtube videos and metadata
function download_videos {
    printf "Downloading videos...\n"
    printf "BATCH_FILE set to %s\n" "${BATCH_FILE}"
    printf "DOWNLOAD_ARCHIVE_FILE set to %s\n" "${DOWNLOAD_ARCHIVE_FILE}"
    printf "DOWNLOAD_VIDEOS_LOG_FILE_PATH set to %s\n" "${DOWNLOAD_VIDEOS_LOG_FILE_PATH}"
    :> "${DOWNLOAD_VIDEOS_LOG_FILE_PATH}"
    YOUTUBE_LIBRARY_COUNT_START=$(wc -l < "${INVENTORY_FILE}")
    # check the batch file for malformed urls
    # test for bad video urls

    # initialize temporary channel urls files for newly added channels during this run
    printf -v CHANNEL_URLS_LIST_TMP_FILE "%s/%s" "${LISTS_DIR}" "yt-dlp-new-channel-urls.txt"
    :> "${CHANNEL_URLS_LIST_TMP_FILE}"
    # Start the video download process and log the output to the download videos log file
    printf "Before download: Your Youtube library has %s videos.\n" "${YOUTUBE_LIBRARY_COUNT_START}"
    # set +eo pipefail
    # Download episodes
    caffeinate "${YTDLP_FILE}" \
        --add-metadata \
        --batch-file "${BATCH_FILE}" \
        "${COOKIE_ARGS[@]}" \
        --download-archive "${DOWNLOAD_ARCHIVE_FILE}"  \
        --write-subs \
        --sub-langs en \
        --embed-subs \
        --embed-chapters \
        --compat-options allow-unsafe-exec-expansion \
        --ffmpeg-location "${FFMPEG_LOCATION}"  \
        --exec "echo '%(channel_url)s' | grep -v -f ${CHANNEL_LIST_URLS} >> ${CHANNEL_LIST_URLS} || true"  \
        --exec "echo '%(channel_url)s' | grep -v -f ${CHANNEL_LIST_URLS} >> ${CHANNEL_URLS_LIST_TMP_FILE} || true" \
        --exec "${TOUCH_FILE} -t '%(upload_date)s0000' '%(filepath)s'" \
        --force-overwrites  \
        --ignore-no-formats-error \
        --js-runtimes "deno:/opt/homebrew/bin/deno" \
        --js-runtimes "node:/opt/homebrew/bin/node" \
        --js-runtimes "quickjs:/opt/homebrew/bin/qjs" \
        --mark-watched  \
        --match-filters "is_live != true" \
        --match-filters "was_live != true" \
        --match-filters "live_status!=is_upcoming" \
        --match-filters "availability!=subscriber_only" \
        --match-filters "original_url!*=/shorts/ & url!*=/shorts/"  \
        --no-abort-on-error  \
        --no-playlist  \
        --no-progress \
        --no-quiet  \
        --no-warnings \
        --no-write-playlist-metafiles \
        --output "%(channel)s [%(channel_id)s]/Season %(upload_date>%Y)s/%(channel)s - S%(upload_date>%Y)sE%(upload_date>%m%d)s%(timestamp>%H%M%S)s - %(title)s [%(id)s].%(ext)s" \
        --paths "home:${HOME_DIR}" \
        --paths "temp:${TEMP_DIR}"  \
        --remux-video mkv \
        --replace-in-metadata channel,title "/" "⧸" \
        --replace-in-metadata channel,title "\?" "" \
        --replace-in-metadata channel,title ":" "-" \
        --replace-in-metadata channel,title "\*" "" \
        -t "mkv" \
        -t sleep \
        --verbose \
        --write-info-json \
        --embed-thumbnail \
        --write-thumbnail \
        --convert-thumbnails jpg \
        &> "${DOWNLOAD_VIDEOS_LOG_FILE_PATH}"
    # set -eo pipefail    
    # update the inventory file after the download is complete
    generate_inventory_file
    # count the number of videos in the library after the download is complete
    YOUTUBE_LIBRARY_COUNT_END=$(wc -l < "${DOWNLOAD_ARCHIVE_FILE}")
    printf "Post download: Your Youtube library has %s videos. %s were added in this run.\n\n" " ${YOUTUBE_LIBRARY_COUNT_END}" "$(( YOUTUBE_LIBRARY_COUNT_END - YOUTUBE_LIBRARY_COUNT_START ))"
}

# Downloads video metadata only for all existing videos
function download_video_metadata {
    printf "Downloading video file metadata\n"
    printf "BATCH_FILE set to %s\n" "${BATCH_FILE}"
    printf "DOWNLOAD_ARCHIVE_FILE set to %s\n" "${DOWNLOAD_ARCHIVE_FILE}"
    printf "DOWNLOAD_VIDEO_METADATA_LOG_FILE_PATH set to %s\n" "${DOWNLOAD_VIDEO_METADATA_LOG_FILE_PATH}"
    # clear the log file before starting the download
    :> "${DOWNLOAD_VIDEO_METADATA_LOG_FILE_PATH}"

    caffeinate "${YTDLP_FILE}" \
        --add-metadata \
        --batch-file "${BATCH_FILE}" \
        "${COOKIE_ARGS[@]}" \
        --write-subs \
        --sub-langs en \
        --embed-subs \
        --embed-chapters \
        --compat-options allow-unsafe-exec-expansion \
        --ffmpeg-location "${FFMPEG_LOCATION}"  \
        --exec "echo '%(channel_url)s' | grep -v -f ${CHANNEL_LIST_URLS} >> ${CHANNEL_LIST_URLS} || true" \
        --exec "${TOUCH_FILE} -t '%(upload_date)s0000' '%(filepath)s'" \
        --force-overwrites  \
        --js-runtimes "deno:/opt/homebrew/bin/deno" \
        --js-runtimes "node:/opt/homebrew/bin/node" \
        --js-runtimes "quickjs:/opt/homebrew/bin/qjs" \
        --mark-watched  \
        --ignore-no-formats-error \
        --match-filters "is_live != true" \
        --match-filters "was_live != true" \
        --match-filters "live_status!=is_upcoming" \
        --match-filters "availability!=subscriber_only" \
        --match-filters "original_url!*=/shorts/ & url!*=/shorts/"  \
        --no-abort-on-error  \
        --no-playlist  \
        --no-progress \
        --no-quiet  \
        --no-warnings \
        --no-write-playlist-metafiles \
        --output "%(channel)s [%(channel_id)s]/Season %(upload_date>%Y)s/%(channel)s - S%(upload_date>%Y)sE%(upload_date>%m%d)s%(timestamp>%H%M%S)s - %(title)s [%(id)s].%(ext)s" \
        --paths "home:${HOME_DIR}" \
        --paths "temp:${TEMP_DIR}"  \
        --remux-video mkv \
        --merge-output-format mkv \
        --replace-in-metadata channel,title "/" "⧸" \
        --replace-in-metadata channel,title "\?" "" \
        --replace-in-metadata channel,title ":" "-" \
        --replace-in-metadata channel,title "\*" "" \
        --skip-download \
        -t sleep \
        --verbose \
        --write-info-json \
        --write-thumbnail \
        --convert-thumbnails jpg \
        &> "${DOWNLOAD_VIDEO_METADATA_LOG_FILE_PATH}"
    printf "Video file metadata downloaded!\n\n"
}

# download channel metadata including thumbnails
function download_channel_metadata {
    # set the batch file to be used in this run
    printf "Downloading channel metadata for %s...\n" "${HOME_DIR}"
    printf "BATCH_FILE set to %s\n" "${CHANNEL_URLS_LIST_TMP_FILE:-CHANNEL_LIST_URLS}"
    printf "DOWNLOAD_ARCHIVE_FILE set to %s\n" "${DOWNLOAD_ARCHIVE_FILE}"
    printf "DOWNLOAD_CHANNEL_METADATA_LOG_FILE_PATH set to %s\n" "${DOWNLOAD_CHANNEL_METADATA_LOG_FILE_PATH}"
    :> "${DOWNLOAD_CHANNEL_METADATA_LOG_FILE_PATH}"

    caffeinate "${YTDLP_FILE}" \
        --batch-file "${CHANNEL_URLS_LIST_TMP_FILE:-CHANNEL_LIST_URLS}" \
        --compat-options allow-unsafe-exec-expansion \
        "${COOKIE_ARGS[@]}" \
        --write-thumbnail \
        --convert-thumbnails jpg \
        --write-info-json \
        --no-abort-on-error  \
        --paths home:"${HOME_DIR}" \
        --paths temp:"${TEMP_DIR}" \
        "${REPLACE_IN_METADATA_ARGS[@]}" \
        --skip-download \
        -t sleep \
        --verbose \
        --output "%(channel)s [%(channel_id)s]/%(title)s [%(channel_id)s]" \
        --playlist-items '0,0' \
        &> "${DOWNLOAD_CHANNEL_METADATA_LOG_FILE_PATH}"
    printf "Channel metadata for %s downloaded!\n\n" "${HOME_DIR}"
    
}

# generates video and channel NFO files from info.json file data
function generate_nfo_files {
    printf "Generating NFO files for %s...\n" "${HOME_DIR}"
    make
    ./bin/yt-dlp-utils nfo --root-dir "${HOME_DIR}" 
    printf "NFO files generated!\n\n"
}

# Updates the list of unavailable videos in the library
function update_unavailable_video_file {
    if [ ! -f "${DOWNLOAD_VIDEOS_LOG_FILE_PATH}" ]; then
        printf "Error log file %s does not exist.\n\n" "${DOWNLOAD_VIDEOS_LOG_FILE_PATH}"
        return 1
    fi
    # Extract unavailable video error messages from the error log
    unavailable_video_files="$(grep -E "ERROR:.*" "${DOWNLOAD_VIDEOS_LOG_FILE_PATH}" | sort -k 4 | cut -d' ' -f2,3,4-  | sed 's/[][]//g; s/://g')"   
    unavailable_video_files="$(grep -v "rate-limited" <<< "${unavailable_video_files}" || true)"   
    unavailable_video_files="$(grep -v "Please sign in." <<< "${unavailable_video_files}" || true)"   
    unavailable_video_files="$(grep -v "Sign in to confirm your age." <<< "${unavailable_video_files}" || true)"   
    unavailable_video_files="$(grep -v "HTTP Error 404" <<< "${unavailable_video_files}" || true)"   
    # add the unavailable video files to the UNAVAILABLE_FILE
    if [ -n "${unavailable_video_files}" ]; then
        printf "%s\n" "${unavailable_video_files}" >> "${VIDEO_UNAVAILABLE_FILE}"
        # make sure the VIDEO_UNAVAILABLE_FILE is unique and sorted
        sort -o "${VIDEO_UNAVAILABLE_FILE}" -u "${VIDEO_UNAVAILABLE_FILE}"
        sort -k 3 -o "${VIDEO_UNAVAILABLE_FILE}" "${VIDEO_UNAVAILABLE_FILE}"
        printf "Unavailable video files updated in %s\n\n" "${VIDEO_UNAVAILABLE_FILE}"
    else
        printf "No unavailable video files found in the error log.\n\n"
    fi        
}

# Start the main loop

# Change cwd to the script's parent directory.
cd "$(dirname "$(dirname "$0")")" || exit 1

# check required resources are set and/or exist
parse_arguments "$@"
load_env
set_cookie_cmd
check_required_variables
check_required_directories
check_required_files
initialize_log_file
update_ytdlp

case "${ACTION}" in
    download-videos)
        printf "Action '%s' selected.\n" "${ACTION}"
        generate_inventory_file
        generate_download_archive
        set_batch_file 
        if ! validate_batch_file; then
            printf "Error: Setting batch file failed. Exiting...\n"
            exit 1
        fi
        if ! cleanup_temp_dir; then
            printf "Error: Cleaning up temp dir failed. Exiting...\n"
            exit 1
        fi
        if ! download_videos; then
            printf "Error: Downloading videos failed. Exiting...\n"
            exit 1
        fi
        if ! cleanup_temp_dir; then
            printf "Error: Cleaning up temp dir failed. Exiting...\n"
            exit 1
        fi
        if ! download_channel_metadata; then
            printf "Error: Downloading channel metadata failed. Exiting...\n"
            exit 1
        fi
        if ! cleanup_temp_dir; then
            printf "Error: Cleaning up temp dir failed. Exiting...\n"
            exit 1
        fi
        if ! generate_nfo_files; then
            printf "Error: Downloading channel nfo files. Exiting...\n"
            exit 1
        fi
        update_unavailable_video_file
        ;;
    download-video-metadata)
        printf "Action '%s' selected.\n" "${ACTION}"
        generate_inventory_file
        generate_download_archive
        if [ ! -z "${WANTED_HOME_DIR}" ]; then 
            generate_wanted_list
        fi
        set_batch_file
        if ! validate_batch_file; then
            printf "Error: Setting batch file failed. Exiting...\n"
            exit 1
        fi

        if ! cleanup_temp_dir; then
            printf "Error: Cleaning up temp dir failed. Exiting...\n"
            exit 1
        fi
        download_video_metadata
        if ! cleanup_temp_dir; then
            printf "Error: Cleaning up temp dir failed. Exiting...\n"
            exit 1
        fi
        if ! generate_nfo_files; then
            printf "Error: Downloading channel nfo files. Exiting...\n"
            exit 1
        fi
        ;;
    download-channel-metadata)
        printf "Action '%s' selected.\n" "${ACTION}"
        set_batch_file
        if ! validate_batch_file; then
            printf "Error: Setting batch file failed. Exiting...\n"
            exit 1
        fi
        if ! cleanup_temp_dir; then
            printf "Error: Cleaning up temp dir failed. Exiting...\n"
            exit 1
        fi

        if ! download_channel_metadata; then
            printf "Error: Downloading channel metadate failed. Exiting...\n"
            exit 1
        fi
        if ! cleanup_temp_dir; then
            printf "Error: Cleaning up temp dir failed. Exiting...\n"
            exit 1
        fi
        if ! generate_nfo_files; then
            printf "Error: Downloading channel nfo files. Exiting...\n"
            exit 1
        fi
        ;;
    generate-channel-urls)
        printf "Action '%s' selected.\n" "${ACTION}"
        generate_channel_urls_file
        ;;
    update-unavailable-videos)
        printf "Action '%s' selected.\n" "${ACTION}"
        update_unavailable_video_file
        ;;
    *)  
        printf "Action not %s not recognized. Exiting...\n" "${ACTION}"
        print_help
        exit 1
        ;;
esac




