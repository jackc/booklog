require "test_helper"

class BookCrudTest < IntegrationTest
  parallelize_me!

  def test_book_crud_cycle
    browser.goto "#{session.app_host}/user_registration/new"
    browser.text_field(label: "Username").set "test"
    browser.text_field(label: "Password").set "secret phrase"
    browser.button(text: "Sign up").click
    assert browser.a(text: "New Book").exist?

    browser.a(text: "New Book").click
    browser.text_field(label: "Title").set "Paradise Lost"
    browser.text_field(label: "Author").set "John Milton"
    browser.date_field(label: "Finish Date").set "2019-01-01"
    browser.text_field(label: "Media").set "audio"
    browser.button(text: "Save").click
    assert browser.dd(text: "Paradise Lost").exist?

    browser.a(text: "Edit").click
    browser.text_field(label: "Title").set "Paradise Regained"
    browser.button(text: "Save").click
    assert browser.dd(text: "Paradise Regained").exist?

    browser.a(text: "Delete").click
    browser.button(text: "Delete").click
    assert browser.a(text: "New Book").exist?
    refute browser.a(text: "Paradise Regained").exist?
  end
end
