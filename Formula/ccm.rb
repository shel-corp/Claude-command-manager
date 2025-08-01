class Ccm < Formula
  desc "TUI for managing Claude Code custom commands"
  homepage "https://github.com/sheltontolbert/claude_command_manager"
  url "https://github.com/sheltontolbert/claude_command_manager/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "" # Will be calculated when creating actual release
  head "https://github.com/sheltontolbert/claude_command_manager.git", branch: "main"
  
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