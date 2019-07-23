require "test_helper"
require "bcrypt"

class LoginTest < IntegrationTest
  parallelize_me!

  def test_user_login_and_logout
    session.db[:users].insert username: "test", password_digest: BCrypt::Password.create("secret phrase")

    browser.goto "#{session.app_host}/login"
    browser.text_field(label: "Username").set "test"
    browser.text_field(label: "Password").set "secret phrase"
    browser.button(text: "Login").click
    assert browser.button(text: "Logout").exist?
  end
end
