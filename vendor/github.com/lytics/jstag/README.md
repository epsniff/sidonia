JSTag - Javascript Analytics Collector Tag
===============================================

A very simple open-source javascript tag for collecting events from a browser to send to a server.  (like google analytics, but name/value pairs defined by you) 


Simple inline implementation:
----------------------------------
```html
    <!--[if lt IE 8]><script src="/static/shim.js"></script><![endif]-->
    <script src="/js/io.js" type="text/javascript" />
    <script type="text/javascript">
      $(document).ready(function(e){
        jstag.init({ cid: 1}).send({myid:1234,category:'books'});
      })
    </script>
```

Async implementation.
-------------------------
This is better for performance, non-blocking page usage.   Use the http://github.com/lytics/jstag/async.js tag as a template and copy/paste (with edits) into a script block on page (do not reference the file or else the value of async is removed).  


```html
    <script type="text/javascript">
      window.jstag=function(e){var t=!1,n=window,r=document,i="/static/io",s=Array.prototype.slice,o=e.url||"";return n.jstag||{load:function(){var e,s=r.getElementsByTagName("script")[0];return t=!0,"JSON"in n&&Array.prototype.forEach||(i+="w"),r.getElementById(i)?this:(e=r.createElement("script"),e.id=i,e.src=o+i+".min.js",s.parentNode.insertBefore(e,s),this)},_q:[],_c:e,bind:function(e){this._q.push([e,s.call(arguments,1)])},ready:function(){this._q.push(["ready",s.call(arguments)])},send:function(){return t||this.load(),this._q.push(["ready","send",s.call(arguments)]),this},ts:(new Date).getTime()}
      }({cid:"CUSTOMER_ID",url:"//collector.domain.com"})
      .send({category:"hello"});// this send is purely optional, it will send as soon as 
      // the tag is loaded
    </script>

    <a href="#" id="testlink" >test link</a>
    <script type="text/javascript" charset="utf-8">
      $(document).ready(function(){
        $("#testlink").click(function(){
          jstag.send({event:"adding_fb_post",conversion:"posting"});
          return false;
        });
      });
    </script>
```

Configuration Options
-------------------------
* *cid* Collection ID, info is posted to */c/cid* allowing 
* *cookie* Name of the Cookie for userid
* *stream* send data to */c/cid/streamname* 


Advanced usage for event bindings. 
--------------------------------------
Often when using a tag, you have a single *Include* of tag, and you have different portion's of your site, or different javascript libraries that need to collect different data.  In that situation, it is easy to utilize the event libraries.  
```html
    <script type="text/javascript">
      // async js tag include (not shown)
    </script>

    <script type="text/javascript" src="/js/ads.js">
      // lets maybe add information from our ad library
      jstag.bind("send.before",function(o) {
        o.data["my_id"] = "value"
        o.data["category"] = "value2"
      })
    </script>
```

Data Format
-----------------
The data is formatted to name=value& format that can be used in querystrings, or form submission.   There are a couple specific formatting issues: 
*  nested objects are flattened to period seperated name=value pairs
*  arrays are sent as custom format

```html
    <script type="text/javascript">
      jstag.send({user:{id:22,name:"aaron"}})
      // would be sent as 
      // user.id=22&user.name=aaron
      jstag.send({user:{id:22,group:["admin","api"]}})
      // would be sent as 
      // user.id=22&user.group=[admin,api]
    </script>
```

Development & Hacking
---------------------------

See 