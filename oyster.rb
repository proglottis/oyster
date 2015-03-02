require "formula"

class Oyster < Formula
  homepage "https://github.com/proglottis/oyster"
  head "https://github.com/proglottis/oyster.git"
  url "https://github.com/proglottis/oyster/archive/v0.1.1.zip"
  sha1 "c1846d07d7db1715e13b4b1cb72600889935bff4"

  depends_on "go" => :build

  def install
    ENV["GIT_DIR"] = cached_download/".git" if build.head?
    ENV["GOBIN"] = bin
    ENV["GOPATH"] = buildpath
    ENV["GOHOME"] = buildpath
    system "go", "get"
    system "go", "build", "-o", "oyster"
    bin.install "oyster"
    bash_completion.install "oyster_bash_completion.sh"
  end

  plist_options :manual => "oyster server"

  def plist; <<-EOS.undent
    <?xml version="1.0" encoding="UTF-8"?>
    <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
    <plist version="1.0">
      <dict>
        <key>Label</key>
        <string>#{plist_name}</string>
        <key>ProgramArguments</key>
        <array>
          <string>#{opt_bin}/oyster</string>
          <string>server</string>
        </array>
        <key>RunAtLoad</key>
        <true/>
        <key>KeepAlive</key>
        <true/>
        <key>StandardErrorPath</key>
        <string>#{var}/log/oyster.log</string>
        <key>StandardOutPath</key>
        <string>#{var}/log/oyster.log</string>
      </dict>
    </plist>
    EOS
  end

  test do
  end
end
