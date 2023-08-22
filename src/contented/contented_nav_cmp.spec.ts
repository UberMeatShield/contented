import {Subscription} from 'rxjs';
import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';

import {RouterTestingModule } from '@angular/router/testing';
import {ContentedNavCmp} from '../contented/contented_nav_cmp';
import {Container} from '../contented/container';

import {ApiDef} from '../contented/api_def';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';
import {GlobalNavEvents, NavTypes} from '../contented/nav_events';

import * as _ from 'lodash';
import * as $ from 'jquery';
import {MockData} from '../test/mock/mock_data';

describe('TestingContentedNavCmp', () => {
    let fixture: ComponentFixture<ContentedNavCmp>;
    let service: ContentedService;
    let comp: ContentedNavCmp;
    let el: HTMLElement;
    let de: DebugElement;
    let httpMock: HttpTestingController;
    let sub: Subscription;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule, ContentedModule, HttpClientTestingModule],
            providers: [
                ContentedService
            ]
        }).compileComponents();

        service = TestBed.inject(ContentedService);
        httpMock = TestBed.inject(HttpTestingController);
        fixture = TestBed.createComponent(ContentedNavCmp);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.contented-nav-cmp'));
        el = de.nativeElement;
    });

    afterEach(() => {
        if (sub) {
            sub.unsubscribe();
            sub = null;
        }
    });

    it('Should create a contented component', () => {
        expect(comp).toBeDefined("We should have the Contented comp");
        expect(el).toBeDefined("We should have a top level element");
    });

    it('Should be able to handle certain key events', () => {
        let counter = 0;
        sub = GlobalNavEvents.navEvts.subscribe(evt => {
            counter++;
        });

        comp.handleKey('w');
        comp.handleKey('a');
        comp.handleKey('s');
        comp.handleKey('d');

        comp.handleKey('e');
        comp.handleKey('q');
        comp.handleKey('x');
        comp.handleKey('f');
        fixture.detectChanges();


        fixture.whenStable().then(() => {
            expect(counter).toEqual(8, "It should have handled these events");
        });
    });

    it("Should be able to handle a document keypress", fakeAsync(() => {
        fixture.detectChanges();

        let validate: NavTypes = null;
        sub = GlobalNavEvents.navEvts.subscribe(evt => {
            validate = evt.action;
        });

        document.dispatchEvent(new KeyboardEvent("keypress", {key: "a"}));
        fixture.detectChanges();

        fixture.whenStable().then(() => {
            expect(validate).toEqual(NavTypes.PREV_MEDIA);
        });
        fixture.detectChanges();
        tick(2000);
        tick(2000);
    }));

});

