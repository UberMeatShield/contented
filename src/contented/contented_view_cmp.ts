import {OnInit, OnDestroy, Component, EventEmitter, Input, Output, HostListener} from '@angular/core';
import {Content} from './content';
import {GlobalNavEvents, NavTypes} from './nav_events';
import {Subscription} from 'rxjs';
import {Screen} from './screen';
import {ContentedService} from './contented_service';
import * as _ from 'lodash';

@Component({
    selector: 'contented-view',
    templateUrl: './contented_view.ng.html'
})
export class ContentedViewCmp implements OnInit, OnDestroy {

    @Input() content: Content;
    @Input() forceWidth: number;
    @Input() forceHeight: number;
    @Input() visible: boolean = false;
    @Input() showScreens = true;
    @Input() restrictContentId = "";

    public maxWidth: number;
    public maxHeight: number;
    public sub: Subscription;

    // This calculation does _not_ work when using a dialog.  Fix?
    // Provide a custom width and height calculation option
    constructor(public _service: ContentedService) {

    }

    public ngOnInit() {
        this.sub = GlobalNavEvents.navEvts.subscribe(evt => {
            switch(evt.action) {
                case NavTypes.VIEW_FULLSCREEN:
                    let content = evt.content || this.content;
                    if (this.restrictContentId && content && content.id !== this.restrictContentId) {
                        return;
                    }
                    console.log("Viewscreen show content", this.content);
                    this.visible = true;
                    this.content = content;
                    if (this.content) {
                        this.scrollContent(this.content);
                        this.handleTextContent(this.content);
                    }
                    break;
                case NavTypes.HIDE_FULLSCREEN:
                    console.log("Viewscreen hide content");
                    if (this.visible && this.content) {
                        GlobalNavEvents.scrollContentView(this.content);          
                    }
                    this.visible = false;
                    break;
                case NavTypes.SELECT_MEDIA:
                    this.content = evt.content;
                    break;
            }
            // console.log("Listen for the fullscreen");
        }) ;
        this.calculateDimensions();

        // I don't like this hack but from a visibility standpoint it kinda works.
        if (this.content && this.content.isText() && this.visible) {
            this._service.getTextContent(this.content).subscribe(
                text => {
                    this.content.fullText = text;
                },
                console.error
            );
        }
    }

    openWindow(content: Content) {
        window.open(content?.fullUrl);
    }

    public ngOnDestroy() {
        this.sub && this.sub.unsubscribe();
    }

    public handleTextContent(content: Content) {
        // This would be better in a method but I would like another example type.
        if (this.content.isText() && !this.content.fullText) {
            this._service.getTextContent(content).subscribe(
                (text: string) => {
                    content.fullText = text;
                    // TODO: _uxLoading
                    // TODO: _uxLanguageGuess (file extention lookups)
                }, 
                console.error
            )
        }
    }

    @HostListener('window:resize', ['$event'])
    public calculateDimensions() {
        // Probably should just set it via dom calculation of the actual parent
        // container?  Maybe?  but then I actually DO want scroll in some cases.
        if (this.forceWidth > 0) {
            this.maxWidth = this.forceWidth;
        } else {
            let width = window.innerWidth; // 
            this.maxWidth = width - 20 > 0 ? width - 20 : 640;
        }

        if (this.forceHeight > 0) {
            // Probably need to do this off the current overall container
            console.log("Force height", this.forceWidth, this.forceHeight);
            this.maxHeight = this.forceHeight;
        } else {
            let height = window.innerHeight; // document.body.clientHeight;
            this.maxHeight = height - 20 > 0 ? height - 64 : 480;
        }
    }

    public scrollContent(content: Content) {
        _.delay(() => {
            let id = `MEDIA_VIEW`;
            let el = document.getElementById(id)
            if (el) {
                el.scrollIntoView(true);
                window.scrollBy(0, -30);
            }
        }, 10);
    }

    public clickedScreen(evt) {
        let seconds = evt.screen.parseSecondsFromScreen();
        let videoEl = <HTMLVideoElement> document.getElementById(`VIDEO_${this.content.id}`);
        videoEl.currentTime = seconds;
    }
}
