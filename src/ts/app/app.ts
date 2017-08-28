import {Component, OnInit} from '@angular/core';
import {Http} from '@angular/http';

@Component({
    selector: 'contented-app',
    template: require('./app.ng.html')
})
export class App implements OnInit {
    public title = "Contented";

    constructor(private http: Http) {

    }

    ngOnInit() {
    }
}

