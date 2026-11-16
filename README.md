## Digital Autonomy
This repository is the 3rd Year Dissertation for the course BSc Computer Science with Industrial Experience at The University of Manchester.

The project aims to provide 'assisted financial management for the digital age'. Older adults, alongside other vulnerable user groups, often are stripped of financial independence due to the evergrowing dependence on online banking applications. This is a factor of digital exclusion - and the number of older adults is continually growing. 

Digital Autonomy builds on the Open-Banking API by introducing a proxy delegating middleware layer to address existing gaps: allowing trusted carers to assist and monitor without the primary user losing autonomy.

Student ID: 11343751

# Services
- api: The implementation of the middleware layer proposed.
- payment-api: A stateless mock interface for the Open Banking API standard to handle payments.
- account-api: A stateless mock interface for the Open Banking API standard to access account details.
- event-notification-api: A stateless mock interface for the Open Banking API standard to handle event notifications.
- db: The Postgres DB instance used to provide state management for this prototype.
- adminer: An open-source UI for providing a web-UI for the database state.

# Architecture
Coming Soon...
<!-- - diagram and explanation of how the API, database, adn mock open banking services fit gother -->


# Endpoints and Example Responses

Coming soon...
<!-- TODO: Implement -->

# Testing
Coming Soon...

# Limitations
This project is limited by the fact that access to the production Open Banking API currently in use is infeasible - as this requires FCA approval. We handle this by using a stateless mock API provided via Mockoon which has been generated against the Open Banking Open-API specification. This means that the project has to create an internal state of 'banking' to track user accounts and ensure simulations are working as expected. In reality, this aspect of the project would not be needed when integrated to the wider Open Banking ecosystem.
