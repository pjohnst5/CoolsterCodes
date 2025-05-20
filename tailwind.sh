curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.15/tailwindcss-macos-arm64
mv tailwindcss-macos-arm64 tailwindcss
chmod +x tailwindcss
./tailwindcss -i ./content/stylesheets/tailwind_base.css -o ./content/stylesheets/tailwind.css
./tailwindcss -i ./content/stylesheets/tailwind_base.css -o ./content/stylesheets/tailwind.min.css --minify
rm tailwindcss
