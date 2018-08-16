#!/usr/bin/python
import sys, string, json, random

def weighted_json_choice(choices):
   '''dict where key is choice and value probability'''
   total = sum(choices[choice] for choice in choices)
   r = random.uniform(0, total)
   upto = 0 
   for choice in choices:
      if upto + choices[choice] >= r:
         return choice
      upto += choices[choice]

def gen_body_data(size):
        return "".join([random.choice(string.ascii_letters + string.digits) for i in xrange(size)])

def make_ammo(method, url, headers_dict, case, body):
    """ makes phantom ammo """
    headers_list = ['{}: {}'.format(header, item) for header, item in headers_dict.iteritems()]
    headers = '\r\n'.join(headers_list)

    req_template = ( 
        "%s %s HTTP/1.1\r\n"
        "%s\r\n"
        "\r\n"
    )   

    req_template_w_entity_body = ( 
        "%s %s HTTP/1.1\r\n"
        "%s\r\n"
        "Content-Length: %d\r\n"
        "\r\n"
        "%s\r\n"
        "\r\n"
    )   

    if method == 'GET' or method == 'DELETE':
        req = req_template % (method, url, headers)
    else:
        req = req_template_w_entity_body % (method, url, headers, len(body), body)

    ammo_template = ( 
        "%d %s\n"
        "%s"
    )   

    return ammo_template % (len(req), case, req)


def choose_case(profile, body):
    case = weighted_json_choice(profile)
    entry = {}
    if case == 'case1':
        entry['case'] = case
        entry['body'] = ''
        entry['method'] = 'GET'
        entry['url'] = '/case1'
    if case == 'case2':
        entry['case'] = case
        entry['body'] = body
        entry['method'] = 'POST'
        entry['url'] = '/case2'
    return entry

if __name__ == '__main__':
    body = gen_body_data(1024)
    try:
        with open ('./raw.ammo', 'w') as ammofile:
            for i in range(10):
                entry = choose_case(json.loads(sys.argv[1]), body)
                ammofile.write(make_ammo(
                    method=entry['method'],
                    url=entry['url'],
                    headers_dict={
                        'Host': 'test.yourhost.net',
                        'Content-Type': 'text/html',
                        'Accept': '*/*',
                        'Connection': 'keep-alive'
                    },
                    case=entry['case'],
                    body=entry['body']
                ))
    finally:
        ammofile.close()
