// https://docs.cypress.io/api/introduction/api.html

describe('My First Test', () => {
  beforeEach(() => {
    cy.log('I run before every test in every spec file!!!!!!')
  })
  it('visits the app root url', () => {
    cy.visit('/')
    cy.visit('http://localhost:8082/login')
    cy.contains('h1', '一二三四五').click()

    cy.url()
    // cy.url().should('include', '/register')
    cy.get('.ant-menu-title-content')
    cy.visit('/login')
    // cy.request('http://localhost:6666/api/atricle')
    // cy.getCookie('your-session-cookie').should('exist')

    cy.get('.password')
      .type('fake@email.com')
      .should('have.value', 'fake@email.com')
    cy.get('.username')
      .type('2021552555555')
      .should('have.value', '2021552555555').click()
    cy.contains('登录').click()
    cy.get('.btn-login').click()
    cy.log('I run before every test in every spec file!!!!!!')
  })
})
