require "minitest/autorun"
require "minitest/reporters"
require "watir"
require "pry"
require "yaml"
require "erb"
require "sequel"

Minitest::Reporters.use! [Minitest::Reporters::DefaultReporter.new(:color => true)]

class Session
  attr_reader :app_host, :http_server_pid, :db

  def initialize(release_func, session_config)
    @release_func = release_func

    @http_server_pid = Process.spawn session_config.fetch("command_name"),
      *session_config["command_args"],
      out: session_config["stdout"],
      err: session_config["stderr"]

    @app_host = session_config.fetch("app_host")

    @db = Sequel.connect session_config.fetch("database_url")
    db[:finished_book].delete
    db[:login_account].delete
  end

  def release
    @release_func.call self
  end
end

class SessionPool
  def initialize(config)
    @all_sessions = config["sessions"].map do |s|
      Session.new self.method(:release), s
    end

    @available_sessions = Queue.new
    @all_sessions.each { |s| @available_sessions.push s }
  end

  def acquire
    @available_sessions.pop
  end

  def close
    @available_sessions.close

    @all_sessions.each do |session|
      Process.kill "TERM", session.http_server_pid
    end
  end

private

  def release(session)
    @available_sessions.push(session)
  end

end

test_config = YAML.load(ERB.new(File.read("test/config.yml")).result)
$session_pool = SessionPool.new(test_config)

Minitest.after_run { $session_pool.close }



class IntegrationTest < Minitest::Test
  def setup
    super
    @session = $session_pool.acquire
    @browser = Watir::Browser.new
  end

  def teardown
    @browser.close
    @session.release
    super
  end

  def session
    @session
  end

  def browser
    @browser
  end
end
