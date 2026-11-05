# Requires: ImageMagick (convert), WebP tools (cwebp)
# Install ImageMagick: https://imagemagick.org/script/download.php
# Install WebP tools: https://developers.google.com/speed/webp/download
# Usage: Run this script in PowerShell from the project root (or adjust paths as needed)

$src = "e:\oss\libr\Libr\core\mod_client\frontend\src\components\assets\icon_bg.png"
$dstRoot = "e:\oss\libr\Libr\mobile-app\android\app\src\main\res"

$targets = @(
    @{ folder = "mipmap-mdpi";    size = 48 },
    @{ folder = "mipmap-hdpi";    size = 72 },
    @{ folder = "mipmap-xhdpi";   size = 96 },
    @{ folder = "mipmap-xxhdpi";  size = 144 },
    @{ folder = "mipmap-xxxhdpi"; size = 192 }
)

foreach ($t in $targets) {
    $outPng = "$dstRoot\$($t.folder)\ic_launcher.png"
    $outWebp = "$dstRoot\$($t.folder)\ic_launcher.webp"
    magick convert $src -resize $($t.size)x$($t.size) $outPng
    cwebp -q 100 $outPng -o $outWebp
    Remove-Item $outPng
    Write-Host "Updated $outWebp with $($t.size)x$($t.size) icon."
}

Write-Host "All launcher icons updated!"
