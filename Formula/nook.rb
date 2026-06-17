class Nook < Formula
  desc "Workspace organizer CLI for developers"
  homepage "https://github.com/lorenzo-vecchio/nook"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/lorenzo-vecchio/nook/releases/download/v0.1.0/nook-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/lorenzo-vecchio/nook/releases/download/v0.1.0/nook-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/lorenzo-vecchio/nook/releases/download/v0.1.0/nook-linux-arm64.tar.gz"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/lorenzo-vecchio/nook/releases/download/v0.1.0/nook-linux-amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  def install
    bin.install "nook"
  end

  test do
    system "#{bin}/nook", "--version"
  end
end
