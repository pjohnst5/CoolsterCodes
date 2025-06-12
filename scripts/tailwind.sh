curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.15/tailwindcss-macos-arm64
mv tailwindcss-macos-arm64 tailwindcss
chmod +x tailwindcss
./tailwindcss -i ./web/stylesheets/tailwind_base.css -o ./web/stylesheets/tailwind.css
./tailwindcss -i ./web/stylesheets/tailwind_base.css -o ./web/stylesheets/tailwind.min.css --minify
rm tailwindcss
