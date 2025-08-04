class Ccm < Formula
  desc "TUI for managing Claude Code custom commands"
  homepage "https://github.com/shel-corp/Claude-command-manager"
  url "https://github.com/shel-corp/Claude-command-manager/archive/refs/tags/v0.0.1.tar.gz"
  sha256 "1f01a523677ec7fb47046cbecdc16d5c8b1db3c369bf6f7a8edf97a1bed9e210" # Will be calculated when creating actual release
  head "https://github.com/shel-corp/Claude-command-manager.git", branch: "main"
  
  # Uncomment when license is added to repository
  # license "MIT"

  depends_on "go" => :build

  def install
    # Build the main binary with custom name
    system "go", "build", *std_go_args(ldflags: "-s -w", output: bin/"ccm"), "./cmd"
  end

  test do
    # Test that the binary was installed and responds to help command
    assert_match "Claude Command Manager", shell_output("#{bin}/ccm help")
    
    # Test that version information is available
    assert_match "Usage:", shell_output("#{bin}/ccm help")
    
    # Test that the binary can handle invalid commands gracefully
    assert_match "Unknown command", shell_output("#{bin}/ccm invalid-command", 1)
  end
end