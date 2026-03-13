require "rake/clean"
require "fileutils"

CLOBBER.include("build")

file "build/frontend/.vite/manifest.json" => [*FileList["css/*.css"]] do
  sh "vite build"
end

# This task is for development convenience - it builds the binary for the current platform and has the convenient code
# for debugging (no optimizations, no inlining). The build matrix tasks below are for CI and release builds.
file "build/booklog" => FileList["Rakefile", "*.go", "go.*", "**/*.go"].exclude(/_test.go$/) do |t|
  sh "go build -o build/booklog"
  # To enable debugging
  # sh %q[go build -o build/booklog -gcflags="all=-N -l"]
end

# Build matrix: all OS/arch combinations
BUILD_TARGETS = [
  {os: "linux", arch: "amd64"},
  {os: "linux", arch: "arm64"},
  {os: "darwin", arch: "amd64"},
  {os: "darwin", arch: "arm64"},
].freeze

GO_SOURCES = FileList["Rakefile", "*.go", "go.*", "**/*.go"].exclude(/_test.go$/)
HTML_SOURCES = FileList["html/**/*.html"]

# Generate file tasks for each target
BUILD_TARGETS.each do |target|
  dir = "build/#{target[:os]}_#{target[:arch]}"
  ext = target[:os] == "windows" ? ".exe" : ""
  binary = "#{dir}/booklog#{ext}"
  html_dir = "#{dir}/html"
  frontend_dir = "#{dir}/frontend"

  # Binary depends on Go sources
  file binary => GO_SOURCES do |t|
    mkdir_p dir
    sh "GOOS=#{target[:os]} GOARCH=#{target[:arch]} go build -o #{t.name}"
  end

  # HTML copy depends on source templates
  file "#{html_dir}/.copied" => HTML_SOURCES do |t|
    rm_rf html_dir
    cp_r "html", html_dir
    touch t.name
  end

  # Frontend depends on Vite build
  file "#{frontend_dir}/.copied" => "build/frontend/.vite/manifest.json" do |t|
    rm_rf frontend_dir
    cp_r "build/frontend", frontend_dir
    touch t.name
  end

  # VERSION file with git commit hash
  version_file = "#{dir}/VERSION"
  task version_file do |t|
    mkdir_p dir
    commit = `git rev-parse HEAD`.chomp
    dirty = `git status --porcelain`.strip.empty? ? "" : "-dirty"
    File.write(t.name, "#{commit}#{dirty}\n")
  end

  # Convenience task for full build directory
  desc "Build artifact for #{target[:os]}/#{target[:arch]}"
  task dir => [binary, "#{html_dir}/.copied", "#{frontend_dir}/.copied", version_file]

  # Tarball of the release directory
  tarball = "#{dir}.tar.gz"
  file tarball => dir do |t|
    sh "tar -czf #{t.name} -C #{dir} ."
  end
end

desc "Build"
task build: ["build/booklog", "build/frontend/.vite/manifest.json"]

desc "Run booklog"
task run: :build do
  exec "build/booklog serve --dev"
end

desc "Watch for source changes and rebuild and rerun"
task :rerun do
  exec %q[watchexec -r -f Rakefile -f "bee/**" -f "cmd/**" -f "data/**" -f "server/**" -f "route/**" -f "validate/**" -f "view/**" -- rake run]
end

file "tmp/test/.databases-prepared" => FileList["postgresql/**/*.sql", "test/testdata/*.sql"] do
  sh "psql -f test/setup_test_databases.sql > /dev/null"
  sh "touch tmp/test/.databases-prepared"
end

desc "Perform all preparation necessary to run tests"
task "test:prepare" => [:build, "tmp/test/.databases-prepared"]

desc "Run all tests"
task test: "test:prepare" do
  sh "go test ./..."
  sh "npx playwright test --config test/playwright/playwright.config.ts"
end

desc "Run Playwright browser tests"
task "test:playwright" => "test:prepare" do
  sh "npx playwright test --config test/playwright/playwright.config.ts"
end

task default: :test
