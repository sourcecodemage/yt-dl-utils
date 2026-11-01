#!/bin/bash
set -eo pipefail


# convert-to-mkv Converts leftover webm files to mp4 and moves the  webm file.
# this is a fix for previous failed conversions.
# Find webm and mp4 files and convert them to mkv format
function convert-to-mkv {
    # -dept 2 
    "${FIND_FILE}" "${HOME_DIR}" -depth 2 -type f -iname "*.webm" -o -iname "*.mp4" -print0 | while IFS= read -r -d '' file; do
        out="${file%.*}.mkv"
        printf "Converting ${file} to ${out}\n"
        # -nostdin to prevent ffmpeg from eating some of the input. This prevents the beginning of the file path from being removed which causes
        # file not found errors.
        # -sn  removes subtitles which mkv does not support.
        # -y overwrite files withough asking.
        if "${FFMPEG_FILE}" -i "${file}" -c copy -sn -nostdin -y "${out}"; then
            touch -r "${file}" "${out}"
            printf "File %s converted to mkv file %s.\n" "${file}" "${out}"
            rm -f "${file}"
            printf "File %s deleted.\n" "${file}" 
        fi
    done
}

function check_required_env_variables {

    required_env_vars = (
        "HOME_DIR"
        "FIND_FILE"
        "FFMPEG_FILE"
    )

    # Loop through the array of variable names
    for var in "${required_env_vars[@]}"; do
        # ${!var} uses indirect expansion to check the value of the variable named by $var
        if [[ -z "${!var}" ]]; then
            echo "Error: Required variable '$var' is not set or is empty."
            ((missing_count++))
        fi
    done

    # Exit if any variable was missing
    if (( missing_count > 0 )); then
        echo "Initialization failed. Missing $missing_count required argument(s)."
        exit 1
    fi

    echo "All required arguments are set successfully!"

}



script_filename="${0##*/}"
script_name="${script_file%.*}"

ENV_FILE=$script_name
# Source env file. Throw error if it does not exist.
if [ -f ${ENV_FILE} ]; then
    source ${ENV_FILE}
else
    printf "Env file %s not found. Exiting..." "${ENV_FILE}"
    exit 1
fi


check_required_env_variables
convert-to-mkv

