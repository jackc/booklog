require "test_helper"
require "bcrypt"

class BookSecurityTest < IntegrationTest
  parallelize_me!

  def test_books_index_redirects_anonymous_users_to_login
    user_id = session.db[:users].insert username: "test", password_digest: BCrypt::Password.create("secret phrase")
    session.db[:books].insert user_id: user_id, title: "Foo", author: "Bar", finish_date: Date.new(2019,1,1), media: "book"

    browser.goto "#{session.app_host}/users/test/books"
    assert_equal "#{session.app_host}/login", browser.url
  end

  def test_books_index_is_forbidden_to_other_users
    user_id = session.db[:users].insert username: "test", password_digest: BCrypt::Password.create("secret phrase")
    session.db[:books].insert user_id: user_id, title: "Foo", author: "Bar", finish_date: Date.new(2019,1,1), media: "book"

    other_user_id = session.db[:users].insert username: "other", password_digest: BCrypt::Password.create("secret phrase")

    browser.goto "#{session.app_host}/login"
    browser.text_field(label: "Username").set "other"
    browser.text_field(label: "Password").set "secret phrase"
    browser.button(text: "Login").click
    assert browser.button(text: "Logout").exist?

    browser.goto "#{session.app_host}/users/test/books"
    assert_match /Forbidden/, browser.text
  end
end
