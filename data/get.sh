#!/bin/bash

# see this https://urs.earthdata.nasa.gov/documentation/for_users/data_access/curl_and_wget

total=$(wc -l < srtm30m_urls.txt)
count=0
MAX_JOBS=300   # max concurrent downloads

# function to count background jobs
function wait_for_jobs() {
  while (( $(jobs -rp | wc -l) >= MAX_JOBS )); do
    sleep 1
  done
}

while IFS= read -r url || [[ -n "$url" ]]; do
  wait_for_jobs  # wait if max jobs running

  wget -q --load-cookies ~/.urs_cookies --save-cookies ~/.urs_cookies --keep-session-cookies "$url" &

  ((count++))
  # Print progress on the same line
  printf "\rProgress: %4d / %d URLs downloaded" "$count" "$total"

  if (( count % 100 == 0 )); then
    sleep 5
  fi
done < srtm30m_urls.txt

wait  # wait for all background jobs to finish

echo -e "\nDone!"

