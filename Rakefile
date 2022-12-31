begin
  require "bundler"
  Bundler.setup
rescue LoadError
  puts "You must `gem install bundler` and `bundle install` to run rake tasks"
end

require "rake/clean"
require "fileutils"
require "rake/testtask"

CLOBBER.include("build", "view/*.html.go")

directory "build/static/css"

file "build/static/css/main.css" => ["build/static/css", *FileList["css/*.scss"]] do
  sh "node-sass --output-style compresses css/main.scss > build/static/css/main.css"
end

file "build/booklog" => [*FileList["**/*.go"]] do |t|
  sh "go build -o build/booklog"
end

file "build/booklog-linux" => [*FileList["**/*.go"]] do |t|
  sh "GOOS=linux GOARCH=amd64 go build -o build/booklog-linux"
end

html_views = Rake::FileList.new("view/*.html")
task view: html_views.ext(".html.go")

rule(/view\/.*\.html\.go$/ => [ proc { |f| f.sub(/\.go$/, "") } ]) do |t|
  sh "gel < #{t.prerequisites.first} | goimports > #{t.name}"
end

desc "Build"
task build: [:view, "build/booklog", "build/static/css/main.css"]

desc "Run booklog"
task run: :build do
  exec "build/booklog serve --insecure-dev-mode"
end

desc "Watch for source changes and rebuild and rerun"
task :rerun do
  exec "react2fs -dir cmd,css,data,server,route,validate,view rake run"
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
Rake::TestTask.new(:test) do |t|
  t.libs << "test"
  t.test_files = FileList['test/**/*_test.rb']
  t.warning = false # Watir generates a lot of warnings.
end

task default: :test
