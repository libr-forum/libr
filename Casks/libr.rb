cask "libr" do
  version "1.0.0"
  sha256 "d710c4612702660afe6c7f94f0467f45d1898f74363e5dcbc37f2909bd5d67b4"

  url "https://github.com/cherry-aggarwal/LIBR/releases/download/v#{version}/libr.dmg"
  name "libr"
  desc "A powerful library manager for macOS"
  homepage "https://example.com"

  livecheck do
    url "https://github.com/cherry-aggarwal/LIBR/releases"
    strategy :github_latest
  end

  depends_on macos: ">= :catalina"

  app "libr.app"

  zap trash: [
    "~/Library/Application Support/libr",
    "~/Library/Preferences/com.libr.app.plist"
  ]
end