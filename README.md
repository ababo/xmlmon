# XML Monitor

Library for monitoring changes in dynamic XML-documents.

## Motivation

There are lots of applications, services and even hardware devices which dynamically generate XML-documents describing current state of their operation. Sometimes it's desirable to track such changes being able to reconstruct the monitored state for any desired moment of the past or to produce different kinds of reports with analytics, diagrams, etc.

## Capabilities

Currently this library is capable to:

- add document schemas by analyzing provided XSD-files
- add documents for a previously created scheme
- to commit updated documents (internally storing structured diff between current and previous versions)
- checkout documents for any previous commit (by specifying a timestamp)

## Installation

1. Install PostgreSQL.
2. Install Google Go language toolchain.
3. Run `go get github.com/lib/pq`
4. Paste your database connection string into `config.json`.
5. Call `mon.Install` function to create a database layout needed by the library.

## Usage

To add a new document schema:

1. Create an XSD-file for your XML-document sample using your favorite XSD-generator (there are several reasonably good ones online). If your generator supports different styles of the resulting schemas prefer nested inline types over flat named ones.

2. Verify correctness of the generated schema. Fix it if needed, but before try to find another generator. Prefer generators which automatically identify integer types, otherwise you'll need to specify that manually.

	<sup><sub>**Note:** Some XSD-generators produce totally incorrect schema for some kinds of samples; others merge types for elements with a same name but different document paths, which will produce extra storage redundancy.

3. For each element with potentially multiple entries in your XML-document find its `element` definition in the generated XSD-schema and add in there a `MonId` attribute. Its value should be the name of an attribute of the original element which uniquely identifies it among other siblings.

	For example, we have the following XML-document:
	
	```xml
	<element1>
		<element2 attr1="value1">
			<element3>
				<element4 attr2="value2">value3</element4>
				<element4 attr2="value4">value5</element4>
				<element4 attr2="value6">value7</element4>
			</element3>
		</element2>
		<element2 attr1="value8">
			<element3>
				<element4 attr2="value9">value10</element4>
				<element4 attr2="value11">value12</element4>
			</element3>
		</element2>
	</element1>
	```
	
	There are at least two multi-entry elements:
	- `/element1/element2`
	- `/element1/element2/element3/element4`

	This means we need to add `monId` attributes into the corresponding XSD `element` definitions:

	> &lt;xs:element name="element2" maxOccurs="unbounded" minOccurs="0" **monId="attr1"**&gt;

	>&lt;xs:element name="element4" maxOccurs="unbounded" minOccurs="0" **monId="attr2"**&gt;

4. Use `mon.AddSchema` function to create an internal schema representation.

Now you can make subsequent document updates using `mon.Commit` function, as well as to reconstruct it using `mon.Checkout` function.

## Limitations

1. Only a narrow subset of XSD specification is yet supported (though it's quite sufficient for most of the cases).
2. Only document reconstruction is yet supported (neither analysis nor visualization).
3. Only string and integer storage types are yet supported for document attributes and content.
