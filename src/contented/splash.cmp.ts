/**
 * Provide the ability to edit the descriptions of content and containers.  Also provide the ability
 * to quickly manage tags.
 */
let resume = `
Justin Carlson 
"Senior <Full-Stack|Lead> Software Engineer" # Old Code Monkey
justinc4@gmail.com

About.
========================================================================================
Likes: Technical challenges, intense coding sessions, competent staff, and management 
that can get me a feature list not written by a lobotomized nepotistic rodent.

Dislikes: eternally cycling process meetings, precise 'estimates' or deathmarch
scoping sessions about how a shirt size didn't match the task.

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
    Amazon Web Services (EC2, Open Search, RDS, SQS, S3, Route53)
    Azure (Mostly Authentication and role management)
    Atlassian JIRA Cloud & Server management
  Databases:
    MySQL, Postgres
    Oracle & The Hell of Driver Management
  DevOps:
    Ansible, Terraform, GitLab CI setups and the inevitable shell script

Experience.
========================================================================================
(2015, 2023) "Senior Full-Stack Engineer"  # Secureworks
  Managed the development of the main web interface into countermeasure lifecycle and creation.
  Built out a web scraping platform that aids in scanning for software vulnerabilities.
  Coded a workflow that provides rapid authoring of intelligence for our Analyst team.
  Coded APIs and UI that displayed malware information to our researchers.
  Upgraded and maintained where malware and virus countermeasures are authored.
  Wrote Ansible playbooks to build out and deploy our services and tweaked Teraform.
  * Typescript (Angular.io), Python, AWS, MySQL, Ansible, Docker, Gitlab CI, Azure

(2013, 2015) "Senior Frontend Engineer"   # Learning Objects
  Created question editors and student views for a serious of education platforms.
  Created course-ware editing tools and display along with account management.
  Process improvements like code review policies, testing requirements and PM guides.
  Tested js code using Karma/Jasmine along with service testing with Ruby
  * Javascript, Java

(2012, 2013) "Frontend Software Developer"  # American Institute for Research
  Created a complex equation editor for mathematical testing.
  Designed widgets and tools for accessible tests for K-12 Education.
  * Javascript, MS IIS

(2010, 2012) "Lead Software Developer"  # Thermopylae Sciences and Technology
  A core contributor for the iSpatial library.
  Lead the development of many projects using our underlying iSpatial library.
    Helped communicate deadlines and reasonable expectations to clients.
  Implemented Javascript libraries around Google Earth and Google Maps.
    Unified the display of data (your point, tracker, polygon to look similar
    on maps or earth).
  Provided tracking functionality for mobile devices and commercial trackers.
  Mentored a host of young programmers and tried to clean up company web coding.
  * Javascript, PHP, Postgres

(2005, 2010) "Software Development Engineer"  # Amazon.com: CS Apps
  Designed and launched improvements that cut page loads in half.
  Lead the modernization of platform JavaScript libraries and design principles.
    Reduced load times by doing less work up front and ajax loading content.
  Ownership of the Concession & Audit libraries and management UI.
    Experienced with SOX rules and regulations along with PCI compliance.
  Implemented a distributed system handling mass processing of amazon issues.
    Reduced contacts & help enable an entire team to be repurposed for other work.
  Automated our builds and transferred ownership from SDEs to support teams.
    Helped optimize a monthly release down to biweekly production releases.
  * Perl, Javascript, Java

(2004, 2005) "Embedded Systems Engineer"  # Boeing: Phantom Works Division
   Developed an RSA crypto library for an embedded system & Math Lib
   Awarded for technological demonstration at US Northcom.
   * C

(2003, 2004)  "Software Test Engineer"  # Microsoft: Sustained Engineering
  Setup multiple root CA authorities and various different domain setups.
  Helped run Crypto tests and setup certificate and signing networks.
  Assisted in running the Active Directory test passes.
  * Active Directory setups and C#

(2002, 2003)  "Software Dev Intern"  # CHILL: National Weather Weather Facility
  Site: http://www.chill.colostate.edu/w/CSU_CHILL
  Dev Lead on project to develop a modern UI displaying Radar weather data.
  Managed an engineering intern program at CSU and mentored several students.
  * Javascript, Java

(2001, 2002)  "Software Dev Intern"  # Los Alamos Advanced Computing Laboratory
  Helped produce a distributed image cache for the LANL MPI program.
  LANL MPI setup and development and assisted in project setups and testing.
  * C

Education.
========================================================================================
B.S. in Computer Science from Colorado State University, Graduation 2003
Minor in Mathematics
`;
import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild} from '@angular/core';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, NgForm, FormControl, FormGroup, Validators} from '@angular/forms';
import {finalize, debounceTime, distinctUntilChanged} from 'rxjs/operators';
import {MatRipple} from '@angular/material/core';
import {EditorComponent} from 'ngx-monaco-editor-v2';
import {ContentedService} from './contented_service';
import {Content} from './content';

import * as _ from 'lodash-es';

@Component({
  selector: 'splash-cmp',
  templateUrl: './splash.ng.html',
})
export class SplashCmp implements OnInit {

  @ViewChild('EDITOR') editor?: EditorComponent;

  @Input() editForm?: FormGroup;
  @Input() editorValue: string = resume;
  @Input() descriptionControl?: FormControl<string>;
  @Input() readOnly: boolean = true;
  @Input() editorOptions = {
    //theme: 'vs-dark',
    //language: 'html',
    language: 'tagging',
  };
  @Input() mc?: Content;

  // These are values for the Monaco Editors, change events are passed down into
  // the form event via the AfterInit and set the v7_definition & suricata_definition.
  @Output() changeEmitter = new EventEmitter<string>();
  public loading: boolean = false;

  // Reference to the raw Microsoft component, allows for
  public monacoEditor?: any;

  constructor(public fb: FormBuilder, public route: ActivatedRoute, public _service: ContentedService) {
  }

  // Subscribe to options changes, if the definition changes make the call
  public ngOnInit() {
    if (!this.editForm) {
      this.editForm = this.fb.group({
        "description": this.descriptionControl = (this.descriptionControl || new FormControl(this.editorValue || "")),
      });
    }
    this.loadSplash();
  }

  // Load the splash page instead of a particular content id
  loadSplash() {
      //this.loading = true;
      console.log("Load splash");
  }


  setReadOnly(state: boolean) {
    this.readOnly = state;
    if (this.monacoEditor) {
      this.monacoEditor.updateOptions({readOnly: this.readOnly});
    }
    if (this.editForm) {
      if (this.readOnly) {
        this.editForm.disable();
      } else {
        this.editForm.enable();
      }
    }
  }

  afterMonacoInit(monaco: any) {
    console.log("Monaco Editor has been initialized");
    this.monacoEditor = monaco;
      // This is a little awkward but we need to be able to change the form control
    if (this.editor) {
      this.changeEmitter.pipe(
        distinctUntilChanged(),
        debounceTime(250)
      ).subscribe(
        (val: string) => {
            console.log("Changed", val);
        },
        console.error
      );
      this.editor.registerOnChange((val: string) => {
        this.changeEmitter.emit(val);
      });
    }
    this.afterMonaco();
    //this.setReadOnly(this.readOnly);
  }

  public afterMonaco() {
    if (!this.editForm) {
      return;
    }

    // Subscribes specifically to the description changes.
    let control = this.editForm.get("description");
    if (control) {
      control.valueChanges.pipe(
        distinctUntilChanged(),
        debounceTime(1000)
      ).subscribe(
        (evt: any) => {
          if (this.editForm) {
            console.log(this.editForm);
          }
        }
      );
      this.editorValue = control.value;
    }
    this.fitContent();
  }

  fitContent() {
    console.log("Content fit");
    const container = document.getElementById('SPLASH_FULL');
    let width = container.offsetWidth;
    //container.style.border = '1px solid #f00';
    let ignoreEvent = false;

    const updateHeight = () => {
    	const contentHeight = Math.min(9000, this.monacoEditor.getContentHeight());
    	container.style.width = `${width}px`;
    	container.style.height = `${contentHeight}px`;

      console.log("Container height", container.style.height);
    	try {
    		//ignoreEvent = true;
        setTimeout(() => {

    		  this.monacoEditor.layout({width, height: contentHeight });
        }, 100);
    	} finally {
    		//ignoreEvent = false;
    	}
    };
    // this.monacoEditor.onDidContentSizeChange(updateHeight);
    setTimeout(() => {
      updateHeight();
    }, 100)
  }
}
