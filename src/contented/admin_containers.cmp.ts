import { Component, OnInit } from '@angular/core';
import { Container } from './container';
import { ContentedService } from './contented_service';

// TODO: When styling out the search add a hover and hover text to make it
// more obvious when something can be clicked.
@Component({
  selector: 'admin-container-cmp',
  templateUrl: './admin_containers.ng.html',
})
export class AdminContainersCmp implements OnInit {

    public loading = false;

    constructor(private service: ContentedService) {

    }

    ngOnInit() {

    }
}
