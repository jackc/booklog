require "rake/clean"
require "fileutils"

CLOBBER.include("build")

file "build/frontend/manifest.json" => [*FileList["css/*.css"]] do
  sh "vite build"
end

file "build/booklog" => FileList["Rakefile", "*.go", "go.*", "**/*.go"].exclude(/_test.go$/) do |t|
  sh "go build -o build/booklog"
  # To enable debugging
  # sh %q[go build -o build/booklog -gcflags="all=-N -l"]
end

file "build/booklog-linux" => [*FileList["**/*.go"]] do |t|
  sh "GOOS=linux GOARCH=amd64 go build -o build/booklog-linux"
end

desc "Build"
task build: ["build/booklog", "build/frontend/manifest.json"]

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
end

task default: :test
