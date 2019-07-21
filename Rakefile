begin
  require "bundler"
  Bundler.setup
rescue LoadError
  puts "You must `gem install bundler` and `bundle install` to run rake tasks"
end

require "rake/clean"
require "fileutils"
require "rake/testtask"

CLOBBER.include("build")

directory "build/static/css"

file "build/static/css/main.css" => ["build/static/css", *FileList["css/*.scss"]] do
  sh "node-sass --output-style compresses css/main.scss > build/static/css/main.css"
end

file "build/booklog" => [*FileList["**/*.go"]] do |t|
  sh "go build -o build/booklog"
end

desc "Build"
task build: ["build/booklog", "build/static/css/main.css"]

desc "Run booklog"
task run: :build do
  exec "build/booklog serve --insecure-dev-mode"
end

desc "Watch for source changes and rebuild and rerun"
task :rerun do
  exec "react2fs -dir cmd,css,domain,html,server,validate rake run"
end

desc "Run tests"
task test: :build
Rake::TestTask.new(:test) do |t|
  t.libs << "test"
  t.test_files = FileList['test/**/*_test.rb']
  t.warning = false # Watir generates a lot of warnings.
end

task default: :test
