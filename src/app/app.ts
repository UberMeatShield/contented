import { Component, OnInit } from '@angular/core';
import { Title } from '@angular/platform-browser';
import { ActivatedRoute, Router, NavigationEnd } from '@angular/router';

@Component({
  selector: 'app-contented',
  templateUrl: 'app.ng.html',
  standalone: false,
})
export class App implements OnInit {
  constructor(
    private activatedRoute: ActivatedRoute,
    private router: Router,
    private titleService: Title
  ) {}

  ngOnInit() {
    this.router.events.subscribe(event => {
      if (event instanceof NavigationEnd) {
        const { title } = this.activatedRoute.firstChild?.snapshot.data as { title: string };
        if (title) {
          this.titleService.setTitle(title);
        }
      }
    });
  }
}
