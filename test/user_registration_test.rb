require "test_helper"

class UserRegistrationTest < IntegrationTest
  parallelize_me!

  def test_user_registration
    browser.goto "#{session.app_host}/user_registration/new"
    browser.text_field(label: "Username").set "test"
    browser.text_field(label: "Password").set "secret phrase"
    browser.button(text: "Sign up").click
    assert browser.a(text: "New Book").exist?
  end
end
