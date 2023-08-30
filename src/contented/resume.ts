export const RESUME = `Justin Carlson
"Senior <Full-Stack|Lead> Software Engineer" # Old Code Monkey
justinc4@gmail.com

About.
========================================================================================
I have been a software developer for two decades and have worked on everything from low
level math libraries to front end web services.  I work well as either an individual
contributor or managing teams and scoping features.

Likes: Technical challenges, intense coding sessions, competent staff, and management
that can get me a feature list not written by a lobotomized nepotistic rodent.

Dislikes: eternally cycling process meetings, precise 'estimates' or death march
scoping sessions about how a shirt size didn't match the task

Links.
========================================================================================
https://www.linkedin.com/in/justin-carlson-8943578
https://github.com/UberMeatShield/contented


Technical Skills.
========================================================================================
  Languages:
    Typescript, JavaScript, Python, GoLang, PHP, Ruby, Perl, Java
  Web Development:
    Frontend: Angular.io, Bootstrap, D3, and a little React
    Backend: GoBuffalo, Django, Nginx, Flask, Ruby on Rails
    Design:  Material Design principles in HTML and CSS
  Web Services:
    Amazon Web Services (EC2, OpenSearch, RDS, SQS, S3, Route53)
    Azure (Mostly Authentication and role management)
    Atlassian JIRA Cloud & Server management
  Databases:
    MySQL, Postgres, Oracle
  DevOps:
    Ansible, Terraform, GitLab CI setups and the inevitable shell script

Experience.
========================================================================================
(2015, 2023) "Senior Full-Stack Engineer"  # Secureworks
  Worked in the Counter Threat Unit research organization protecting companies from the
  threat of state actors or individual script kids.  Most of my tasks were scaling out
  UI and backend views dealing with the review of customer threats.
  * Developed and deployed our countermeasure editing and review platform.
    - This is a django based website using Angular and Bootstrap to review changes
    and the activity of current threats.
    - The platform ingested millions of alerts into postgres per hour and provided
    summary graphs using D3.
    - Provided authentication support in Azure and managed access roles with it.
    - This system used Fargate to run clusters of ingestion processes and web frontend.
    - Wrote Terraform code for the creation of Route53 records, RDS setup and docker
    scaling out a cluster.
  * Developed vulnerability ingestion and display platform used in Secureworks.
    - This was a python flask website that managed web scraping for our researchers.
    - The service scraped data into a JIRA instance and provided for data correlation.
    - Built the ansible playbooks that helped with deployments from GitLab CI
  * Created a UI and backend for associating malware with sample detonations.
    - This data was used to help review client vulnerabilities and stored in MySQL
    - Search was provided using AWS OpenSearch(ElasticSearch)

(2013, 2015) "Senior Frontend Engineer"   # Learning Objects
  The company built out competency based learning and test questions for the education
  of students mostly at the college level.
  * Created question editors and views for a series of education platforms.
    - These pages helped with providing competency based learning and grading.
    - This work was mostly an Angular 1 application on a JAVA backend.
  * Developed a modular UI for rendering coursework.

(2012, 2013) "Frontend Software Developer"  # American Institute for Research
  Worked in the group that was tasked with providing test creation tools for writing
  test questions for K-12 education.
  * Created a complex equation editor for mathematical testing in javascript.
    - The math editors were hosted on a Microsoft IIS instance and provided general quiz
    writing.
    - The equation editor was mostly jquery rendered with the popular MathJax library.

(2010, 2012) "Lead Software Developer"  # Thermopylae Sciences and Technology
  Worked in the group that developed geospatial imaging support and tracking for mobile
  devices and other GPS enabled devices.
  * Development on iSpatial which was a geospatial library using Google Earth to render
  assets in near real time for the Department of State.
    - The main backend was written in PHP and the front end was javascript using the
    ExtJS framework (now Sencha).
    - Helped communicate deadlines and reasonable expectations to clients.
  * Implemented Javascript libraries around Google Earth and Google Maps.
    - Unified the display of data (your point, tracker, polygon to look similar
    on maps or earth).
    - Developed tracking functionality for mobile devices and commercial trackers.
  * Mentored young programmers and helped clean up company web coding standards.

(2005, 2010) "Software Development Engineer"  # Amazon.com: CS Apps
  The customer service development environment builds out recall notices, handle
  refund permissions and support calls.
  * Lead the modernization of platform JavaScript libraries and design principles.
    - Reduced load times by doing less work up front and ajax loading content.
    - This Cut page load times on the perl based app from 15+ seconds in India to 2s.
    - Drastically lowered the cost of handling a customer contact.
  * Ownership of the Concession & Audit libraries and management UI.
    - Developed code that conformed to SOX rules and regulations and PCI compliance.
  * Implemented a distributed system handling mass processing of amazon issues.
    - Reduced contacts & help enable an entire team to be repurposed for other work.
    - Wrote JAVA distributed tasks handling large scale order manipulation mailing,
    refunds and fault tolerant jobs.
  * Automated our builds and transferred ownership from SDEs to support teams.
    - Helped optimize a monthly release down to biweekly production releases.

(2004, 2005) "Embedded Systems Engineer"  # Boeing: Phantom Works Division
   * In C developed a RSA crypto library for an embedded system.
     - Awarded for demonstration at US Northcom.

(2003, 2004)  "Software Test Engineer"  # Microsoft: Sustained Engineering
  * Setup multiple root CA authorities and various different domain setups.
    - Helped run Crypto tests and setup certificates and signing networks.
    - Assisted in running the Active Directory test passes.
  * Active Directory setups and writing test code in C#.

(2002, 2003)  "Software Dev Intern"  # CHILL: National Weather Weather Facility
  Worked in weather radar display using Java Swing.
  * Development Lead on project to develop a modern UI displaying Radar weather data.
  * Managed an engineering intern program at CSU and mentored several students.
  Site: http://www.chill.colostate.edu/w/CSU_CHILL

(2001, 2002)  "Software Dev Intern"  # Los Alamos Advanced Computing Laboratory
  * Helped produce a distributed image cache for the LANL MPI program in C.

Education.
========================================================================================
B.S. in Computer Science from Colorado State University, Graduation 2003
Minor in Mathematics
`;
